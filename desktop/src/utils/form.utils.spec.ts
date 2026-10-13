import { FormBuilder, FormGroup } from "@angular/forms";
import { applyApiErrors } from "./form.utils";

const NAME_TAKEN = "an OIDC provider with this name already exists";
const URL_IN_USE =
  "the server public URL cannot be cleared while an OIDC provider is enabled; disable the provider first";

describe("applyApiErrors", () => {
  let form: FormGroup;

  beforeEach(() => {
    form = new FormBuilder().group({ name: "keycloak", serverPublicUrl: "" });
  });

  it("puts the server's message on the control it names", () => {
    applyApiErrors(form, { serverPublicUrl: URL_IN_USE });

    expect(form.get("serverPublicUrl")?.errors).toEqual({ server: URL_IN_USE });
    expect(form.get("name")?.errors).toBeNull();
  });

  // Material's default ErrorStateMatcher shows a mat-error only for a control
  // that is invalid AND touched, so setErrors on its own leaves a field the user
  // never focused looking perfectly fine.
  it("marks the control touched so the error actually renders", () => {
    expect(form.get("name")?.touched).toBe(false);

    applyApiErrors(form, { name: NAME_TAKEN });

    expect(form.get("name")?.touched).toBe(true);
  });

  it("replaces a stale client-side error rather than merging with it", () => {
    form.get("name")?.setErrors({ required: true });

    applyApiErrors(form, { name: NAME_TAKEN });

    expect(form.get("name")?.errors).toEqual({ server: NAME_TAKEN });
  });

  // errorMsg is the shape the HTTP interceptor already toasts. Reporting it a
  // second time would cost the user the message rather than repeat it:
  // MatSnackBar.open() dismisses whatever is showing.
  it("ignores a body the interceptor owns", () => {
    const onUnmatched = jest.fn();

    applyApiErrors(form, { errorMsg: "a server public URL must be configured" }, onUnmatched);

    expect(onUnmatched).not.toHaveBeenCalled();
    expect(form.get("name")?.errors).toBeNull();
    expect(form.get("serverPublicUrl")?.errors).toBeNull();
  });

  it("falls back for a key that names no control", () => {
    const onUnmatched = jest.fn();

    applyApiErrors(form, { files: "at least one image is required" }, onUnmatched);

    expect(onUnmatched).toHaveBeenCalledWith("at least one image is required");
  });

  it("reports several unmatched messages once, not once each", () => {
    const onUnmatched = jest.fn();

    applyApiErrors(form, { files: "need an image", comments: "need a comment" }, onUnmatched);

    expect(onUnmatched).toHaveBeenCalledTimes(1);
    expect(onUnmatched.mock.calls[0][0]).toContain("need an image");
    expect(onUnmatched.mock.calls[0][0]).toContain("need a comment");
  });

  it("stays quiet when every key found a control", () => {
    const onUnmatched = jest.fn();

    applyApiErrors(form, { name: NAME_TAKEN }, onUnmatched);

    expect(onUnmatched).not.toHaveBeenCalled();
  });

  // A failed request does not always carry a parsed object body -- a network
  // failure gives null, and an nginx 502 gives HTML.
  it("does not throw on a body that is not a field map", () => {
    const onUnmatched = jest.fn();

    [null, undefined, "<html>502 Bad Gateway</html>", 42, ["name"]].forEach((body) => {
      expect(() => applyApiErrors(form, body, onUnmatched)).not.toThrow();
    });

    expect(onUnmatched).not.toHaveBeenCalled();
    expect(form.get("name")?.errors).toBeNull();
  });

  it("skips a value that is not a usable message", () => {
    const onUnmatched = jest.fn();

    applyApiErrors(form, { name: "", serverPublicUrl: { nested: "thing" } }, onUnmatched);

    expect(form.get("name")?.errors).toBeNull();
    expect(form.get("serverPublicUrl")?.errors).toBeNull();
    expect(onUnmatched).not.toHaveBeenCalled();
  });
});
