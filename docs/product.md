# Todo app (omi pilot v2)

## Người dùng
Một người dùng, chạy trên máy local. Không đăng nhập.

## Vấn đề
Pilot v2 cho omi: kiểm tra squad xây được một ứng dụng 3 tầng (frontend, backend, database) với PR, review và CI, không chỉ một CLI Python.

## Chức năng v1
1. Thêm một việc (tiêu đề 1-200 ký tự, không rỗng).
2. Xem danh sách việc, mới nhất ở trên.
3. Đánh dấu xong / chưa xong.
4. Xóa một việc.

## Ngoài phạm vi
Đăng nhập, nhiều người dùng, sửa tiêu đề, hạn chót, nhãn, tìm kiếm, phân trang, deploy lên cloud.

## Ràng buộc
- Chạy local bằng `docker compose up`, không hosting, không chi phí.
- Pilot nhanh: ít task, không thêm thư viện khi stdlib làm được.

## Stack
- Frontend: Next.js (App Router, TypeScript) trong `web/` - PO chọn; một trang, gọi API bằng `fetch`.
- Backend: Go, `net/http` của stdlib trong `api/` - PO chọn; router của Go 1.22+ đủ cho 4 endpoint, không cần framework.
- Database: PostgreSQL 16 - PO chọn; driver `pgx`, migration SQL thuần trong `api/migrations/`.
- Kiểm thử: `go test` (domain và HTTP dùng repository giả trong bộ nhớ; test Postgres chỉ chạy khi có `DATABASE_URL`, CI cấp qua service container), `next build` và lint cho web.
- CI: GitHub Actions, job `api` và job `web`.

## API
- `GET /api/todos` -> danh sách
- `POST /api/todos` `{"title": "..."}` -> 201 + việc mới; tiêu đề sai -> 400
- `PATCH /api/todos/{id}` `{"done": true|false}` -> việc đã cập nhật; không có -> 404
- `DELETE /api/todos/{id}` -> 204; không có -> 404

## Tiêu chí thành công
- `docker compose up` rồi mở http://localhost:3000: làm được 4 chức năng; dữ liệu còn sau khi khởi động lại compose.
- CI xanh trên mọi PR; mỗi task đi qua review của omi và gated merge.
- Ghi lại số lần PO phải can thiệp, token theo vai trò và lỗi omi phát hiện (như pilot v1).
