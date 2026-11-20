export const newCounter = (start) => ({ count: start });
export const bump = (counter) => ({ count: counter.count + 1 });
export const count = (counter) => counter.count;
export const label = (text) => (value) => ({ text, value });
export const labelOf = (labelled) => labelled.text;
export const valueOf = (labelled) => labelled.value;
