import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, test, expect } from "vitest";
import App from "./App";

describe("App Component Happy Path", () => {
  test("renders the main page, allows typing and submitting a URL", async () => {
    render(<App />);
    const user = userEvent.setup();

    expect(screen.getByText("Website Analyzer")).toBeInTheDocument();
    const apiKeyInput = screen.getByPlaceholderText("Enter API Key");
    const urlInput = screen.getByPlaceholderText("Enter a website URL");
    const analyzeButton = screen.getByRole("button", { name: /analyze/i });

    await user.type(apiKeyInput, "test-key");
    expect(apiKeyInput).toHaveValue("test-key");

    await user.type(urlInput, "https://new-site.com");
    expect(urlInput).toHaveValue("https://new-site.com");

    await user.click(analyzeButton);

    const tableRow = await screen.findByText("Example Domain");
    expect(tableRow).toBeInTheDocument();
  });
});
