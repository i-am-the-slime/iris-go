export const describePair = (pair) => `A_=${pair.A} A=${pair.a}`;
export const describeTriple = (triple) => `A_=${triple.A} A=${triple.a} A2=${triple.A2}`;
export const describeHeader = (header) =>
  `Content_type2=${header["content-type"]} Content_type=${header.content_type}`;
export const makePair = (upper) => (lower) => ({ A: upper, a: lower });
