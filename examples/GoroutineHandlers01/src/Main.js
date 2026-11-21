export const onTheEventLoop = (effect) => effect;

export const forEach = (values) => (each) => () => values.forEach((value) => each(value)());

export const sortWith = (compare) => (values) => values.slice().sort((a, b) => compare(a)(b));

export const fanOut = (workers) => (calls) => (handler) => (onError, onSuccess) => {
  for (let call = 0; call < workers * calls; call++) handler();
  setTimeout(() => onSuccess(), 0);
  return (cancelError, onCancelerError, onCancelerSuccess) => onCancelerSuccess();
};

export const later = (milliseconds) => (message) => (onError, onSuccess) => {
  setTimeout(() => onSuccess(message), milliseconds);
  return (cancelError, onCancelerError, onCancelerSuccess) => onCancelerSuccess();
};
