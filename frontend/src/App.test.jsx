import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import App from "./App";

describe("App", () => {
  beforeEach(() => {
    globalThis.fetch = vi.fn(() => new Promise(() => {}));
  });

  it("renders the page heading", () => {
    render(<App />);
    expect(
      screen.getByRole("heading", { name: /traveling zip simulator/i })
    ).toBeInTheDocument();
  });
});
