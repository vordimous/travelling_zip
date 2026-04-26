// Generic data-table panel used across the app. Pass `fixedHeight` (number, in
// pixels) to constrain the scroll area; omit it for a free-flowing layout.
export default function DataTable({
  title,
  columns,
  rows,
  emptyMessage,
  fixedHeight,
}) {
  return (
    <section className="panel">
      <div className="panel-header">
        <h2>{title}</h2>
        <span>{rows.length}</span>
      </div>
      {rows.length === 0 ? (
        <p className="empty">{emptyMessage}</p>
      ) : (
        <div
          className="table-wrap"
          style={
            fixedHeight ? { height: fixedHeight, overflowY: "auto" } : undefined
          }
        >
          <table>
            <thead>
              <tr>
                {columns.map((column) => (
                  <th key={column.key}>{column.label}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row, index) => (
                <tr key={row.id ?? `${title}-${index}`}>
                  {columns.map((column) => (
                    <td key={column.key}>{row[column.key]}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
