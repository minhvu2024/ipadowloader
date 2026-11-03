# IPA Downloader

Ứng dụng Windows nhẹ (1 file `.exe`, không cần cài đặt) để **tìm và tải file IPA thẳng từ App Store** bằng Apple ID của bạn. Viết bằng Go + Wails v2 + Tailwind CSS, phần giao tiếp với Apple dùng thư viện [ipatool](https://github.com/majd/ipatool).

File IPA tải về cài được lên mọi iPhone **đăng nhập cùng Apple ID** (qua 3uTools, iMazing, Apple Configurator…), tiện cho máy hay reset.

## Dùng

1. Chạy `IPA Downloader.exe` (cần Microsoft Edge WebView2 — Windows 10/11 có sẵn).
2. Đăng nhập Apple ID → nhập mã 2FA Apple gửi về iPhone/iPad.
3. **Tìm kiếm**: gõ tên app, hoặc dán thẳng bundle ID / app ID. Bấm **Tải IPA**.
   - App miễn phí: tự động "Get" license nếu tài khoản chưa có.
   - App trả phí: phải đã mua bằng chính Apple ID này trên iPhone/App Store trước.
4. **Đang tải**: xem tiến trình, hủy.
5. **Thư viện**: file đã tải (mặc định `%USERPROFILE%\IPA`), đổi thư mục, mở Explorer, xóa.
6. **Tải hàng loạt**: dán danh sách bundle ID (mỗi dòng một cái) → **Kiểm tra** → **Tải tất cả**. Danh sách được lưu để lần sau tải bản cập nhật.

Cài lên iPhone: cắm cáp → 3uTools → *Apps* → *Import & Install* → chọn nhiều file `.ipa` cùng lúc. (3uTools/iMazing đều cần driver Apple — cài app **Apple Devices** từ Microsoft Store là đủ.)

## Dữ liệu lưu ở đâu

- `%APPDATA%\IPADownloader\keychain.json` — token phiên + mật khẩu Apple ID (ipatool cần để tự đăng nhập lại khi token hết hạn). File chỉ nằm trên máy bạn; **nên dùng Apple ID riêng** cho việc tải IPA.
- `%APPDATA%\IPADownloader\cookies.json`, `config.json`
- `<thư mục IPA>\.library.json` — metadata (tên, icon) của file đã tải.

Bấm **Đăng xuất** để xóa token.

## Build từ mã nguồn

Yêu cầu: Go ≥ 1.26, Node.js ≥ 18.

```powershell
.\build.ps1
```

Hoặc dùng Wails CLI: `wails build -platform windows/amd64`.

`go.mod` có khối `replace` trỏ `golang.org/x/*`, `gopkg.in/*`… sang mirror GitHub tương ứng (do môi trường build gốc chỉ truy cập được GitHub). Không ảnh hưởng gì khi build bình thường; muốn bỏ thì xóa khối `replace` rồi `go mod tidy`.

## Cấu trúc

```
main.go          Wails bootstrap, embed frontend/dist
app.go           Các hàm bind cho frontend: Login/Logout/Search/Lookup/Download/Cancel/ListLibrary/...
store.go         Khởi tạo ipatool AppStore: keychain file, cookie jar, machine ID
frontend/
  index.html     Giao diện (Tailwind)
  src/app.js     Logic frontend (vanilla JS), gọi window.go.main.App.*
  src/input.css  Tailwind source → dist/style.css
  dev-mock.js    Backend giả để xem giao diện trong trình duyệt
build/windows/   icon + winres.json (icon/manifest cho exe)
```

Sự kiện Go → JS: `download:progress` (`{id, name, status, percent, downloaded, total, speed, message, path}`), `library:changed`.

## Lưu ý

- IPA từ App Store có mã hóa FairPlay + license gắn với Apple ID: **không** cài được bằng Apple ID khác, cũng không re-sign qua Sideloadly/AltStore được.
- Tải quá nhiều trong thời gian ngắn có thể bị Apple khóa tạm tài khoản.
- Chỉ tải app bạn có quyền dùng.
