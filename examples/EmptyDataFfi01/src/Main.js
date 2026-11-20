export const bag = () => ({ items: [] });
export const put = (item, target) => { target.items.push(item); };
export const contents = (target) => () => target.items.slice();
