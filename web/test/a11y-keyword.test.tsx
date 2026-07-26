/**
 * @jest-environment jsdom
 */
import { render } from "@testing-library/react";
import { axe, toHaveNoViolations } from "jest-axe";
import { KEYWORD_PAGES } from "../lib/seo/keywords";

expect.extend(toHaveNoViolations);

/** Minimal mirror of keyword article markup for axe (server page is async). */
function KeywordArticleFixture({ slug }: { slug: string }) {
  const p = KEYWORD_PAGES.find((k) => k.slug === slug);
  if (!p) throw new Error(`missing keyword ${slug}`);
  return (
    <article lang="vi-VN" className="mx-auto max-w-3xl px-6 py-12">
      <h1>{p.keyword}</h1>
      <p>{p.description}</p>
      {p.intro.map((para) => (
        <p key={para.slice(0, 24)}>{para}</p>
      ))}
      {p.faqs.length > 0 ? (
        <section>
          <h2>Câu hỏi thường gặp</h2>
          {p.faqs.map((f) => (
            <details key={f.q}>
              <summary>{f.q}</summary>
              <p>{f.a}</p>
            </details>
          ))}
        </section>
      ) : null}
      <a href="/login?signup=1">Dùng Shopass miễn phí</a>
    </article>
  );
}

describe("R48 a11y — keyword template", () => {
  it("has no serious axe violations on a sample keyword page", async () => {
    const sample = KEYWORD_PAGES[0];
    const { container } = render(<KeywordArticleFixture slug={sample.slug} />);
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });
});
