import { http, HttpResponse } from "msw";

export const handlers = [
  http.get("/api/analyses", () => {
    return HttpResponse.json([
      {
        id: 1,
        url: "https://example.com",
        title: "Example Domain",
        status: "done",
        internal_links: 1,
        external_links: 0,
        inaccessible_links: 0,
        has_login_form: false,
        h1_count: 1,
        h2_count: 0,
        h3_count: 0,
        h4_count: 0,
        h5_count: 0,
        h6_count: 0,
      },
    ]);
  }),

  http.post("/api/analyze", () => {
    return new HttpResponse(null, { status: 200 });
  }),
];
