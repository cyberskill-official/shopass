# AGENTS.md - SănDeal (slot dành cho giao thức memory CyberOS)

File này được để trống có chủ đích cho **giao thức memory CyberOS (Layer-1 Memory Protocol)** - thứ kích hoạt BRAIN / `.cyberos-memory/`. Hãy thay nội dung file này bằng AGENTS.md của cyberos.

Conventions build SănDeal (cách chọn FR, bất biến không thương lượng, stack, quy ước viết) KHÔNG nằm ở đây - chúng ở [`docs/feature-requests/SHIP-GUIDE.md`](docs/feature-requests/SHIP-GUIDE.md). BACKLOG.md và README.md đã trỏ tới SHIP-GUIDE.md, nên việc thay file này không làm mất conventions.

## Cài AGENTS.md của cyberos (chọn một)

Nguồn: `../cyberos/modules/memory/cyberos/data/AGENTS.md` (giả định shopass và cyberos là thư mục anh em dưới `CyberSkill/`).

Cách 1 - symlink (luôn đồng bộ với cyberos, nhưng cần giữ hai repo cạnh nhau):

```
cd /Users/stephencheng/Projects/CyberSkill/shopass
ln -sf ../cyberos/modules/memory/cyberos/data/AGENTS.md AGENTS.md
```

Cách 2 - copy (độc lập, nhưng phải tự cập nhật khi cyberos đổi giao thức):

```
cp /Users/stephencheng/Projects/CyberSkill/cyberos/modules/memory/cyberos/data/AGENTS.md \
   /Users/stephencheng/Projects/CyberSkill/shopass/AGENTS.md
```

## Kích hoạt memory cho shopass

Giao thức (§0.4) định nghĩa memory-root là `.cyberos-memory/` ở gốc repo. Để BRAIN hoạt động trong shopass, tạo store tại `/Users/stephencheng/Projects/CyberSkill/shopass/.cyberos-memory/` theo layout §2 của giao thức (hoặc trỏ về store dùng chung nếu muốn). Một agent đọc AGENTS.md (giao thức) trước, rồi đọc `docs/feature-requests/BACKLOG.md` để vào việc build; BACKLOG trỏ tiếp tới SHIP-GUIDE.md cho conventions.
