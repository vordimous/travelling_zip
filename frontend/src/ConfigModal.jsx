import { configFields } from "./configSchema";

// Modal form for editing simulation configuration. The parent owns the
// `config` input state (strings) and submits to the backend; this component
// renders fields from a static descriptor list and surfaces per-field errors.
//
// Accessibility: clicking the backdrop closes the modal. Focus trapping and
// Escape-to-close are deferred to post-Step-2.
export default function ConfigModal({
  open,
  config,
  onChange,
  onSubmit,
  onClose,
  error,
  fieldErrors,
}) {
  if (!open) {
    return null;
  }
  return (
    <div
      className="modal-backdrop"
      onClick={onClose}
      role="presentation"
    >
      <div
        className="modal panel"
        role="dialog"
        aria-modal="true"
        aria-label="Edit simulation configuration"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="panel-header">
          <h2>Edit Configuration</h2>
          <button type="button" className="modal-close" onClick={onClose}>
            Close
          </button>
        </div>
        <form className="control-grid" onSubmit={onSubmit}>
          {configFields.map((field) => {
            const { key, label, kind } = field;
            const fieldError = fieldErrors?.[key];
            const value = config[key] ?? "";
            return (
              <label key={key}>
                <span>{label}</span>
                {kind === "enum" ? (
                  <select
                    name={key}
                    value={value}
                    onChange={onChange}
                    aria-invalid={fieldError ? "true" : undefined}
                  >
                    {field.options.map((option) => (
                      <option key={option} value={option}>
                        {option}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    name={key}
                    type="number"
                    min={field.min}
                    step={field.step}
                    value={value}
                    onChange={onChange}
                    aria-invalid={fieldError ? "true" : undefined}
                  />
                )}
                {fieldError ? (
                  <span className="field-error">{fieldError}</span>
                ) : null}
              </label>
            );
          })}
          <button type="submit">Save & Run</button>
        </form>
        {error ? <p className="error">{error}</p> : null}
      </div>
    </div>
  );
}
