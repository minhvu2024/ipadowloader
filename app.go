package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/schollz/progressbar/v3"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------- kiểu dữ liệu trả về cho frontend ----------

type Status struct {
	LoggedIn  bool   `json:"loggedIn"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	OutputDir string `json:"outputDir"`
	Version   string `json:"version"`
}

type LoginResult struct {
	NeedCode bool   `json:"needCode"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Country  string `json:"country"`
}

type AppInfo struct {
	ID        int64   `json:"id"`
	BundleID  string  `json:"bundleId"`
	Name      string  `json:"name"`
	Version   string  `json:"version"`
	Price     float64 `json:"price"`
	Icon      string  `json:"icon"`
	Developer string  `json:"developer"`
	SizeBytes int64   `json:"sizeBytes"`
	Rating    float64 `json:"rating"`
	Genre     string  `json:"genre"`
	MinOS     string  `json:"minOs"`
}

type LibraryItem struct {
	FileName  string    `json:"fileName"`
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	BundleID  string    `json:"bundleId"`
	AppID     int64     `json:"appId"`
	Version   string    `json:"version"`
	SizeBytes int64     `json:"sizeBytes"`
	Icon      string    `json:"icon"`
	ModTime   time.Time `json:"modTime"`
}

type Progress struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Status     string  `json:"status"` // queued | purchasing | downloading | finalizing | done | error | cancelled
	Percent    float64 `json:"percent"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Speed      float64 `json:"speed"` // KB/s
	Message    string  `json:"message"`
	Path       string  `json:"path"`
}

// ---------- App ----------

type App struct {
	ctx   context.Context
	deps  *storeDeps
	mu    sync.Mutex
	jobs  map[int64]context.CancelFunc
	cfg   config
	cfgMu sync.Mutex
}

type config struct {
	OutputDir string `json:"outputDir"`
}

const appVersion = "1.0.0"

func NewApp() *App {
	return &App{jobs: map[int64]context.CancelFunc{}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	deps, err := newStore()
	if err != nil {
		wruntime.LogError(ctx, err.Error())
	}
	a.deps = deps
	a.loadConfig()
}

func (a *App) configPath() string {
	if a.deps == nil {
		return ""
	}
	return filepath.Join(a.deps.Dir, "config.json")
}

func (a *App) loadConfig() {
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	if p := a.configPath(); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(b, &a.cfg)
		}
	}
	if a.cfg.OutputDir == "" {
		home, _ := os.UserHomeDir()
		a.cfg.OutputDir = filepath.Join(home, "IPA")
	}
	_ = os.MkdirAll(a.cfg.OutputDir, 0o755)
}

func (a *App) saveConfig() {
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	if p := a.configPath(); p != "" {
		b, _ := json.MarshalIndent(a.cfg, "", "  ")
		_ = os.WriteFile(p, b, 0o644)
	}
}

func (a *App) outputDir() string {
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	return a.cfg.OutputDir
}

// ---------- tài khoản ----------

func (a *App) account() (appstore.Account, error) {
	if a.deps == nil {
		return appstore.Account{}, errors.New("store chưa khởi tạo")
	}
	info, err := a.deps.Store.AccountInfo()
	if err != nil {
		return appstore.Account{}, err
	}
	return info.Account, nil
}

func (a *App) GetStatus() Status {
	st := Status{OutputDir: a.outputDir(), Version: appVersion}
	acc, err := a.account()
	if err == nil && acc.Email != "" {
		st.LoggedIn = true
		st.Email = acc.Email
		st.Name = acc.Name
		st.Country = countryFromStoreFront(acc.StoreFront)
	}
	return st
}

// Login đăng nhập Apple ID. Lần đầu gọi không có code: nếu tài khoản bật 2FA,
// Apple gửi mã về thiết bị và hàm trả về NeedCode=true. Gọi lại với code để hoàn tất.
func (a *App) Login(email, password, code string) (LoginResult, error) {
	if a.deps == nil {
		return LoginResult{}, errors.New("store chưa khởi tạo")
	}
	email = strings.TrimSpace(email)
	code = strings.TrimSpace(code)
	if email == "" || password == "" {
		return LoginResult{}, errors.New("cần nhập email và mật khẩu")
	}

	out, err := a.deps.Store.Login(appstore.LoginInput{Email: email, Password: password, AuthCode: code})
	if err != nil {
		if errors.Is(err, appstore.ErrAuthCodeRequired) {
			return LoginResult{NeedCode: true, Email: email}, nil
		}
		return LoginResult{}, friendlyErr(err)
	}
	_ = a.deps.CookieJar.Save()

	return LoginResult{
		Email:   out.Account.Email,
		Name:    out.Account.Name,
		Country: countryFromStoreFront(out.Account.StoreFront),
	}, nil
}

func (a *App) Logout() error {
	if a.deps == nil {
		return nil
	}
	err := a.deps.Store.Revoke()
	_ = a.deps.Keychain.Remove("account")
	_ = a.deps.CookieJar.Save()
	if err != nil && !errors.Is(err, errKeyNotFound) {
		return friendlyErr(err)
	}
	return nil
}

// ---------- tìm kiếm ----------

func (a *App) Search(term string, limit int) ([]AppInfo, error) {
	acc, err := a.account()
	if err != nil {
		return nil, errors.New("chưa đăng nhập")
	}
	term = strings.TrimSpace(term)
	if term == "" {
		return []AppInfo{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	out, err := a.deps.Store.Search(appstore.SearchInput{
		Account:  acc,
		Term:     term,
		Limit:    int64(limit),
		Platform: appstore.PlatformIPhone,
	})
	if err != nil {
		return nil, friendlyErr(err)
	}

	res := make([]AppInfo, 0, len(out.Results))
	ids := make([]string, 0, len(out.Results))
	for _, r := range out.Results {
		res = append(res, AppInfo{ID: r.ID, BundleID: r.BundleID, Name: r.Name, Version: r.Version, Price: r.Price})
		ids = append(ids, strconv.FormatInt(r.ID, 10))
	}
	a.enrich(res, ids, countryFromStoreFront(acc.StoreFront))
	return res, nil
}

// Lookup tìm chính xác theo bundle ID hoặc app ID (dùng cho nhập tay / nhập danh sách).
func (a *App) Lookup(query string) (AppInfo, error) {
	acc, err := a.account()
	if err != nil {
		return AppInfo{}, errors.New("chưa đăng nhập")
	}
	query = strings.TrimSpace(query)
	in := appstore.LookupInput{Account: acc, Platform: appstore.PlatformIPhone}
	if id, err := strconv.ParseInt(query, 10, 64); err == nil {
		in.AppID = id
	} else {
		in.BundleID = query
	}
	out, err := a.deps.Store.Lookup(in)
	if err != nil {
		return AppInfo{}, friendlyErr(err)
	}
	r := out.App
	info := []AppInfo{{ID: r.ID, BundleID: r.BundleID, Name: r.Name, Version: r.Version, Price: r.Price}}
	a.enrich(info, []string{strconv.FormatInt(r.ID, 10)}, countryFromStoreFront(acc.StoreFront))
	return info[0], nil
}

// enrich lấy icon / nhà phát triển / dung lượng từ iTunes Lookup API (public, không cần login).
func (a *App) enrich(apps []AppInfo, ids []string, country string) {
	if len(ids) == 0 {
		return
	}
	q := url.Values{}
	q.Set("id", strings.Join(ids, ","))
	q.Set("country", strings.ToLower(country))
	q.Set("entity", "software")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get("https://itunes.apple.com/lookup?" + q.Encode())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))

	var payload struct {
		Results []struct {
			TrackID       int64   `json:"trackId"`
			ArtworkURL100 string  `json:"artworkUrl100"`
			ArtworkURL512 string  `json:"artworkUrl512"`
			SellerName    string  `json:"sellerName"`
			ArtistName    string  `json:"artistName"`
			FileSizeBytes string  `json:"fileSizeBytes"`
			Rating        float64 `json:"averageUserRating"`
			Genre         string  `json:"primaryGenreName"`
			MinOS         string  `json:"minimumOsVersion"`
			Version       string  `json:"version"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return
	}
	byID := map[int64]int{}
	for i, r := range payload.Results {
		byID[r.TrackID] = i
	}
	for i := range apps {
		idx, ok := byID[apps[i].ID]
		if !ok {
			continue
		}
		r := payload.Results[idx]
		icon := r.ArtworkURL100
		if icon == "" {
			icon = r.ArtworkURL512
		}
		apps[i].Icon = icon
		apps[i].Developer = firstNonEmpty(r.SellerName, r.ArtistName)
		apps[i].SizeBytes, _ = strconv.ParseInt(r.FileSizeBytes, 10, 64)
		apps[i].Rating = r.Rating
		apps[i].Genre = r.Genre
		apps[i].MinOS = r.MinOS
		if apps[i].Version == "" {
			apps[i].Version = r.Version
		}
	}
}

// ---------- tải về ----------

func (a *App) emit(p Progress) {
	wruntime.EventsEmit(a.ctx, "download:progress", p)
}

// Download đưa app vào hàng đợi tải. Tiến trình báo qua event "download:progress".
func (a *App) Download(app AppInfo) error {
	acc, err := a.account()
	if err != nil {
		return errors.New("chưa đăng nhập")
	}

	a.mu.Lock()
	if _, running := a.jobs[app.ID]; running {
		a.mu.Unlock()
		return errors.New("app này đang tải")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.jobs[app.ID] = cancel
	a.mu.Unlock()

	a.emit(Progress{ID: app.ID, Name: app.Name, Status: "queued"})

	go func() {
		defer func() {
			a.mu.Lock()
			delete(a.jobs, app.ID)
			a.mu.Unlock()
		}()
		a.runDownload(ctx, acc, app)
	}()
	return nil
}

func (a *App) Cancel(id int64) {
	a.mu.Lock()
	cancel, ok := a.jobs[id]
	a.mu.Unlock()
	if ok {
		cancel()
	}
}

func (a *App) runDownload(ctx context.Context, acc appstore.Account, info AppInfo) {
	store := a.deps.Store
	target := appstore.App{ID: info.ID, BundleID: info.BundleID, Name: info.Name, Version: info.Version}
	outDir := a.outputDir()
	_ = os.MkdirAll(outDir, 0o755)

	fail := func(err error) {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			a.emit(Progress{ID: info.ID, Name: info.Name, Status: "cancelled", Message: "Đã hủy"})
			return
		}
		a.emit(Progress{ID: info.ID, Name: info.Name, Status: "error", Message: friendlyErr(err).Error()})
	}

	// Làm mới token nếu hết hạn (dùng mật khẩu đã lưu khi login).
	relogin := func() error {
		out, err := store.Login(appstore.LoginInput{Email: acc.Email, Password: acc.Password})
		if err != nil {
			return err
		}
		acc = out.Account
		_ = a.deps.CookieJar.Save()
		return nil
	}

	bar := progressbar.NewOptions64(1,
		progressbar.OptionSetWriter(io.Discard),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSetRenderBlankState(false),
	)

	// Goroutine báo tiến trình mỗi 250ms.
	stopTick := make(chan struct{})
	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stopTick:
				return
			case <-t.C:
				s := bar.State()
				a.emit(Progress{
					ID: info.ID, Name: info.Name, Status: "downloading",
					Percent:    s.CurrentPercent * 100,
					Downloaded: int64(s.CurrentBytes),
					Total:      bar.GetMax64(),
					Speed:      s.KBsPerSecond,
				})
			}
		}
	}()

	var out appstore.DownloadOutput
	var lastErr error
	purchased := false

	for attempt := 0; attempt < 4; attempt++ {
		if ctx.Err() != nil {
			lastErr = ctx.Err()
			break
		}
		if errors.Is(lastErr, appstore.ErrPasswordTokenExpired) {
			if err := relogin(); err != nil {
				lastErr = err
				break
			}
		}
		if errors.Is(lastErr, appstore.ErrLicenseRequired) {
			if purchased {
				break
			}
			a.emit(Progress{ID: info.ID, Name: info.Name, Status: "purchasing", Message: "Đang lấy license (Get app miễn phí)…"})
			if err := store.Purchase(appstore.PurchaseInput{Account: acc, App: target, Platform: appstore.PlatformIPhone}); err != nil {
				if errors.Is(err, appstore.ErrPasswordTokenExpired) {
					lastErr = err
					continue
				}
				if !errors.Is(err, appstore.ErrLicenseAlreadyExists) {
					lastErr = err
					break
				}
			}
			purchased = true
		}

		a.emit(Progress{ID: info.ID, Name: info.Name, Status: "downloading"})
		var err error
		out, err = store.Download(appstore.DownloadInput{
			Context:    ctx,
			Account:    acc,
			App:        target,
			OutputPath: outDir,
			Progress:   bar,
			Platform:   appstore.PlatformIPhone,
		})
		if err == nil {
			lastErr = nil
			break
		}
		lastErr = err
		if !errors.Is(err, appstore.ErrPasswordTokenExpired) && !errors.Is(err, appstore.ErrLicenseRequired) {
			break
		}
	}
	close(stopTick)

	if lastErr != nil {
		fail(lastErr)
		return
	}

	a.emit(Progress{ID: info.ID, Name: info.Name, Status: "finalizing", Percent: 100, Message: "Đang ghi license vào file…"})
	if err := store.ReplicateSinf(appstore.ReplicateSinfInput{Sinfs: out.Sinfs, PackagePath: out.DestinationPath}); err != nil {
		fail(err)
		return
	}

	a.saveLibraryMeta(out.DestinationPath, info)
	st, _ := os.Stat(out.DestinationPath)
	var size int64
	if st != nil {
		size = st.Size()
	}
	a.emit(Progress{ID: info.ID, Name: info.Name, Status: "done", Percent: 100, Downloaded: size, Total: size, Path: out.DestinationPath, Message: "Hoàn tất"})
	wruntime.EventsEmit(a.ctx, "library:changed")
}

// ---------- thư viện file đã tải ----------

type libraryMeta map[string]AppInfo // key = file name

func (a *App) libraryMetaPath() string {
	return filepath.Join(a.outputDir(), ".library.json")
}

func (a *App) readLibraryMeta() libraryMeta {
	m := libraryMeta{}
	if b, err := os.ReadFile(a.libraryMetaPath()); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	return m
}

func (a *App) saveLibraryMeta(path string, info AppInfo) {
	a.cfgMu.Lock()
	defer a.cfgMu.Unlock()
	m := libraryMeta{}
	if b, err := os.ReadFile(filepath.Join(a.cfg.OutputDir, ".library.json")); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	m[filepath.Base(path)] = info
	b, _ := json.MarshalIndent(m, "", "  ")
	_ = os.WriteFile(filepath.Join(a.cfg.OutputDir, ".library.json"), b, 0o644)
}

func (a *App) ListLibrary() ([]LibraryItem, error) {
	dir := a.outputDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []LibraryItem{}, nil
	}
	meta := a.readLibraryMeta()
	items := []LibraryItem{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".ipa") {
			continue
		}
		st, err := e.Info()
		if err != nil {
			continue
		}
		it := LibraryItem{FileName: e.Name(), Path: filepath.Join(dir, e.Name()), SizeBytes: st.Size(), ModTime: st.ModTime()}
		if m, ok := meta[e.Name()]; ok {
			it.Name, it.BundleID, it.AppID, it.Version, it.Icon = m.Name, m.BundleID, m.ID, m.Version, m.Icon
		} else {
			it.BundleID, it.AppID, it.Version = parseIPAName(e.Name())
			it.Name = it.BundleID
		}
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModTime.After(items[j].ModTime) })
	return items, nil
}

func (a *App) DeleteFile(fileName string) error {
	fileName = filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(fileName), ".ipa") {
		return errors.New("chỉ xóa được file .ipa")
	}
	if err := os.Remove(filepath.Join(a.outputDir(), fileName)); err != nil {
		return err
	}
	wruntime.EventsEmit(a.ctx, "library:changed")
	return nil
}

func (a *App) ChooseOutputDir() (string, error) {
	dir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            "Chọn thư mục lưu file IPA",
		DefaultDirectory: a.outputDir(),
	})
	if err != nil || dir == "" {
		return a.outputDir(), err
	}
	a.cfgMu.Lock()
	a.cfg.OutputDir = dir
	a.cfgMu.Unlock()
	a.saveConfig()
	_ = os.MkdirAll(dir, 0o755)
	wruntime.EventsEmit(a.ctx, "library:changed")
	return dir, nil
}

func (a *App) OpenOutputDir() error {
	dir := a.outputDir()
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer.exe", dir).Start()
	case "darwin":
		return exec.Command("open", dir).Start()
	default:
		return exec.Command("xdg-open", dir).Start()
	}
}

func (a *App) ShowInFolder(path string) error {
	if runtime.GOOS == "windows" {
		return exec.Command("explorer.exe", "/select,", path).Start()
	}
	return a.OpenOutputDir()
}

// ---------- helpers ----------

func parseIPAName(name string) (bundle string, id int64, version string) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	parts := strings.Split(base, "_")
	if len(parts) >= 3 {
		version = parts[len(parts)-1]
		id, _ = strconv.ParseInt(parts[len(parts)-2], 10, 64)
		bundle = strings.Join(parts[:len(parts)-2], "_")
		return
	}
	return base, 0, ""
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

func friendlyErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case errors.Is(err, appstore.ErrAuthCodeRequired):
		return errors.New("cần mã xác thực 2 lớp (2FA)")
	case errors.Is(err, appstore.ErrPasswordTokenExpired):
		return errors.New("phiên đăng nhập hết hạn, hãy đăng nhập lại")
	case errors.Is(err, appstore.ErrLicenseRequired):
		return errors.New("tài khoản chưa có license cho app này")
	case errors.Is(err, appstore.ErrSubscriptionRequired):
		return errors.New("app này cần mua/đăng ký trước trên iPhone (ứng dụng trả phí)")
	case errors.Is(err, appstore.ErrTemporarilyUnavailable):
		return errors.New("Apple báo tạm thời không khả dụng, thử lại sau")
	case errors.Is(err, appstore.ErrAppNotFound):
		return errors.New("không tìm thấy app trên store của tài khoản này")
	case strings.Contains(msg, "invalid credentials") || strings.Contains(msg, "bad login"):
		return errors.New("sai email hoặc mật khẩu")
	case strings.Contains(msg, "locked"):
		return errors.New("tài khoản đang bị khóa, kiểm tra trên appleid.apple.com")
	}
	return fmt.Errorf("%s", msg)
}

// countryFromStoreFront chuyển mã storefront của Apple (vd "143441-1,29") sang mã quốc gia.
func countryFromStoreFront(sf string) string {
	id := strings.SplitN(sf, "-", 2)[0]
	if c, ok := storeFronts[id]; ok {
		return c
	}
	return "US"
}

var storeFronts = map[string]string{
	"143441": "US", "143462": "JP", "143471": "VN", "143444": "GB", "143455": "CA",
	"143460": "AU", "143443": "DE", "143442": "FR", "143450": "IT", "143454": "ES",
	"143466": "KR", "143465": "CN", "143470": "TW", "143463": "HK", "143464": "SG",
	"143473": "ID", "143474": "MY", "143475": "TH", "143476": "PH", "143467": "IN",
	"143445": "AT", "143446": "BE", "143447": "FI", "143448": "GR", "143449": "IE",
	"143451": "LU", "143452": "NL", "143453": "PT", "143456": "SE", "143457": "NO",
	"143458": "DK", "143459": "CH", "143461": "NZ", "143469": "RU", "143472": "ZA",
	"143468": "MX", "143503": "BR", "143509": "AR", "143480": "TR", "143481": "AE",
	"143483": "SA", "143489": "EG", "143491": "IL", "143478": "PK",
}
