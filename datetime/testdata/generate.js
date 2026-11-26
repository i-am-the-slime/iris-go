// Records purescript-datetime's foreign results for datetime_test.go:
//   bun testdata/generate.js <datetime>/src/Data > testdata/cases.json
const directory = process.argv[2];
const date = await import(directory + "/Date.js");
const dateTime = await import(directory + "/DateTime.js");
const instant = await import(directory + "/DateTime/Instant.js");

let seed = 1915;
const random = () => ((seed = (seed * 1103515245 + 12345) % 2147483648) / 2147483648);
const pick = (items) => items[Math.floor(random() * items.length)];
const number = (n) => (Object.is(n, -0) ? "-0" : String(n));

const years = [-271821, -271820, -1, 0, 1, 4, 99, 100, 1899, 1900, 1969, 1970, 2000, 2024, 9999, 275759, 275760, -2147483648, 2147483647];
const cases = { dates: [], instants: [], toDateTimes: [], diffs: [] };
for (let i = 0; i < 800; i++) {
  const y = Math.max(-2147483648, Math.min(2147483647, pick(years) + (random() < 0.3 ? Math.floor(random() * 5) - 2 : 0)));
  const m = Math.floor(random() * 16) - 1;
  const d = Math.floor(random() * 34) - 1;
  const canonical = date.canonicalDateImpl((a) => (b) => (c) => [a, b, c], y, m, d);
  cases.dates.push({ y, m, d, canonical: canonical.map(number), weekday: number(date.calcWeekday(y, m, d)),
    diff: number(date.calcDiff(y, m, d, pick(years), Math.floor(random() * 14), Math.floor(random() * 32))) });
  const h = Math.floor(random() * 30) - 2, mi = Math.floor(random() * 70) - 3, s = Math.floor(random() * 70) - 3, ms = Math.floor(random() * 1200) - 100;
  cases.instants.push({ y, m, d, h, mi, s, ms, instant: number(instant.fromDateTimeImpl(y, m, d, h, mi, s, ms)) });
  const rec1 = { year: y, month: m, day: d, hour: h, minute: mi, second: s, millisecond: ms };
  const rec2 = { year: pick(years), month: 1 + Math.floor(random() * 12), day: 1 + Math.floor(random() * 28), hour: 3, minute: 4, second: 5, millisecond: 6 };
  cases.diffs.push({ rec1, rec2, diff: number(dateTime.calcDiff(rec1, rec2)) });
}
const instants = [0, -0, 1, -1, 1.5, -1.5, 86399999, -86400001, 8.64e15, -8.64e15, 8.64e15 + 1, 1e300, NaN, 951782400000, -62198755200000, 253402300799999];
for (let i = 0; i < 200; i++) instants.push(Math.floor((random() * 2 - 1) * 8.64e15) + (random() < 0.2 ? 0.25 : 0));
for (const t of instants) {
  cases.toDateTimes.push({ instant: number(t), parts: instant.toDateTimeImpl((a) => (b) => (c) => (d) => (e) => (f) => (g) => [a, b, c, d, e, f, g])(t).map(number) });
}
cases.adjusts = [];
const offsets = [0, -0, 1, -1, 0.5, -0.5, 1.7, -1.7, 86400000, -86400000, 8.64e15, -8.64e15, 1.728e16, -1.728e16, NaN, Infinity, -Infinity];
for (let i = 0; i < 400; i++) {
  const record = { year: pick(years), month: 1 + Math.floor(random() * 12), day: 1 + Math.floor(random() * 31), hour: Math.floor(random() * 24),
    minute: Math.floor(random() * 60), second: Math.floor(random() * 60), millisecond: Math.floor(random() * 1000) };
  const offset = random() < 0.3 ? pick(offsets) : (random() * 2 - 1) * Math.pow(10, Math.floor(random() * 17));
  const adjusted = dateTime.adjustImpl((value) => [value])(null)(offset)(record);
  cases.adjusts.push({ record, offset: number(offset), adjusted: adjusted && adjusted[0] });
}
console.log(JSON.stringify(cases));
