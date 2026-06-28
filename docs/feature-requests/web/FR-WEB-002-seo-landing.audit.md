---
fr_id: FR-WEB-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt

Audit độc lập, tái diễn từ file FR-WEB-002 hiện tại. FR đặc tả tập trang landing SEO cho các cụm keyword GTM §5.7. 12 mệnh đề §1 BCP-14 (11 MUST + 1 SHOULD), testable. Bốn cụm keyword lõi (cách săn xự Shopee, lịch sale, mã freeship, sale thật hay sale ảo) sinh SSG, mỗi trang có metadata riêng + JSON-LD đúng intent (FAQPage/ItemList/Article), sitemap/robots sinh tự động từ một nguồn keywords.ts. Ràng buộc landing công khai (không auth, không gọi API người dùng) giữ funnel liền mạch. Đạt 10/10.

## §2 - Findings

Không còn khiếm khuyết tồn dư. Kiểm độc lập:
- SSG/ISR (§1 #2, DEC-WEB-06) - KHÔNG CSR; crawler đọc HTML response đầu; có test generateStaticParams.
- Metadata đầy đủ qua Metadata API (§1 #3): title/description/canonical/OG/twitter + lang vi-VN; có test seen-set khẳng định title không trùng.
- JSON-LD theo intent (§1 #4): FAQPage cho sale thật hay sale ảo, ItemList cho lịch sale, Article cho guide; có seo-jsonld test.
- sitemap/robots sinh tự động từ keywords.ts (§1 #5, DEC-WEB-09) - không soạn tay; có sitemap test.
- Landing công khai group (marketing) ngoài (app) (§1 #6, DEC-WEB-10) - không auth, không gọi API người dùng; giữ funnel.
- Nội dung trung thực gắn năng lực thật (§1 #10) - tránh spam/cloaking phạt SEO.
- Typography prose plain ASCII + tiếng Việt có dấu; không tự cấm; sentinel có mặt.

## §3 - Bảng truy vết (từ file hiện tại)

| §1 mệnh đề | AC | Test/Artefact |
|---|---|---|
| #1 trang per keyword | #1 | keywords.ts + generateStaticParams |
| #2 SSG/ISR | #1 | route (marketing)/[keyword] |
| #3 metadata đầy đủ | #2,#3 | generateMetadata + seo-metadata test |
| #4 JSON-LD theo intent | #4 | jsonld.ts + seo-jsonld test |
| #5 sitemap/robots tự sinh | #5,#9 | app/sitemap.ts + robots.ts + sitemap test |
| #6 landing công khai | #6 | group (marketing) ngoài matcher |
| #7 CTA funnel | #7 | link đăng ký/cài extension |
| #8 canonical | #2 | alternates.canonical |
| #9 lang vi-VN + h1 | #8 | article lang + h1 |
| #10 nội dung trung thực | - | gắn FR-DEAL-001/FR-WEB-003 |
| #12 tsc/test | #10 | npm test |

## §4 - Kết luận

Mọi mệnh đề normative có mã/test backing; bốn cụm keyword lõi §5.7 đều phủ. SEO render phía server, metadata + JSON-LD đầy đủ, sitemap tự sinh, funnel công khai liền mạch. Không mệnh đề mồ côi. Không cần sửa. Score = 10/10. Verdict: PASS. Sẵn sàng build.

---

*Hết audit FR-WEB-002.*
