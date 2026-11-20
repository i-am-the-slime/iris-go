export const foldPairs = (f) => (initial) => (xs) => xs.reduce((acc, x) => f(acc)(x), initial);

export const pickOperation = (name) => () =>
  name === "times" ? (a) => (b) => a * b : (a) => (b) => a - b;
