/** slugOf turns a tab title into the segment that names it in the URL. Which
    tabs exist is decided at run time, so the URL cannot hold a declared name. */
export function slugOf(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}
