export async function allPages(api, path, token) {
  const values = []; let cursor = null;
  do { const page = await api(path + (cursor ? '?cursor=' + encodeURIComponent(cursor) : ''), 'GET', undefined, token); values.push(...page.data); cursor = page.next_cursor; } while (cursor);
  return values;
}
