// Modal form for editing simulation configuration. The parent owns the
// `config` input state and submits to the backend; this component is the
// presentational shell.
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
          {Object.entries(config).map(([key, value]) => (
            <label key={key}>
              <span>{key}</span>
              <input
                min="0"
                name={key}
                onChange={onChange}
                type="number"
                value={value}
              />
            </label>
          ))}
          <button type="submit">Save & Run</button>
        </form>
        {error ? <p className="error">{error}</p> : null}
      </div>
    </div>
  );
}
