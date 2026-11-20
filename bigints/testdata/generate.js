// Records purescript-bigints' behaviour for bigints_test.go:
//   bun testdata/generate.js <bigints>/src/Data/BigInt.js > testdata/cases.json
// (big-integer must resolve from BigInt.js, as it does in a project that depends on bigints).
// Each result is the value's decimal string, or "THROW <name>: <message>".
const ffi = await import(process.argv[2]);

let seed = 424242;
const random = () => ((seed = (seed * 1103515245 + 12345) % 2147483648) / 2147483648);
const pick = (items) => items[Math.floor(random() * items.length)];
const attempt = (f) => {
  try {
    const value = f();
    return typeof value === "object" && value !== null && typeof value.isNegative !== "boolean" ? value.toString() : JSON.stringify(value);
  } catch (error) {
    return "THROW " + error.name + ": " + error.message;
  }
};
const just = (value) => ({ just: value.toString() });
const nothing = { nothing: true };
const maybe = (value) => (value.just !== undefined ? value.just : "Nothing");

const texts = ["", " ", "0", "-0", "+0", "42", "  42  ", "-42", "+7", "\t\n99\n", " 12 ", "﻿5",
  "0x1F", "0X1f", "0o17", "0b101", "-0x1", "0x", "0b2", "0x1_0", "1_000", "1.5", "1.0", "15e-1", "1e3", "1E3",
  "1e+3", "1e-3", "-1e3", ".5e1", "5.", "1e", "e5", "--5", "-", "+", "abc", "Infinity", "-Infinity", "NaN",
  "9007199254740991", "9007199254740992", "9007199254740993", "-9007199254740993", "123456789012345678901234567890",
  "-123456789012345678901234567890", "1e30", "1.5e30", "-2.25e40", "1e+30", "1E30", "1e30e2", "1.5e0", "12e-1",
  " 1e30", "1e30 ", "1.e30", "0001", "00000000000000000000000000001", "1e21", "1e400", "0x123456789abcdef0123",
  "ff", "FF", "-ff", "zz", "Z", "10", "777", "101", "12z4", "<10>", "1<10>", "<99>0", "<-1>", "<>", "<12", "1-2",
  "-<5>", "İ", "K", "١٢", "1 2", "1\n", "😀", "00", "-00", "1.5e1", "0.5e1", "4294967296e0", "1e3.5", "1e+", "1e-0",
  "123456789abcdefghijklmnopqrstuvwxyz", "1234567890123456789012345678901234567890e5"];
const bases = [10, 2, 8, 16, 36, 1, 0, -1, -2, -10, 37, 100, 3, -16, 62];

const fromBase = [];
for (const text of texts) {
  for (const base of bases) {
    fromBase.push({ base, text, result: maybe(ffi.fromBaseImpl(just)(nothing)(base)(text)) });
  }
}

const numbers = [0, -0, 1.5, -1.5, 0.999, -0.999, 42, 1e21, -1e21, 9007199254740993, 2 ** 64, 1.7976931348623157e308,
  5e-324, NaN, Infinity, -Infinity, 123456789.987, -(2 ** 53)];
const fromNumber = numbers.map((n) => ({ number: Object.is(n, -0) ? "-0" : String(n), result: maybe(ffi.fromNumberImpl(just)(nothing)(n)) }));

const values = ["0", "1", "-1", "2", "-2", "7", "-7", "10", "255", "-255", "1000", "65537", "2147483647", "-2147483648",
  "9007199254740993", "18446744073709551615", "18446744073709551616", "-18446744073709551617",
  "123456789012345678901234567890123", "-123456789012345678901234567890123", "1180591620717411303424"];
for (let i = 0; i < 20; i++) {
  let digits = String(1 + Math.floor(random() * 9));
  for (let j = Math.floor(random() * 60); j > 0; j--) digits += String(Math.floor(random() * 10));
  values.push(random() < 0.5 ? "-" + digits : digits);
}
const big = (text) => ffi.fromTypeLevelInt(text);

const unary = values.map((text) => {
  const x = big(text);
  return {
    value: text,
    toNumber: String(ffi.toNumber(x)),
    abs: attempt(() => ffi.abs(x)),
    even: ffi.even(x),
    odd: ffi.odd(x),
    not: attempt(() => ffi.not(x)),
    bases: bases.concat([5, 7]).map((base) => attempt(() => ffi.toBase(base)(x))),
    digits: [2, 10, 16, 100, -3, 1, 0].map((base) => attempt(() => ffi.digitsInBase(base)(x))),
    shifts: [0, 1, 7, 63, 64, 65, 200, -1, -7, -64, -200, 1.5, NaN, Infinity, 1e7, 10000001, -1e8].map((n) => [attempt(() => ffi.shl(x)(n)), attempt(() => ffi.shr(x)(n))]),
  };
});

const binary = [];
for (const left of values) {
  for (const right of values.slice(0, 24)) {
    const x = big(left), y = big(right);
    binary.push({
      left, right,
      results: [ffi.biAdd, ffi.biSub, ffi.biMul, ffi.biDiv, ffi.biMod, ffi.and, ffi.or, ffi.xor].map((f) => attempt(() => f(x)(y))),
      equals: ffi.biEquals(x)(y),
      compare: ffi.biCompare(x)(y),
    });
  }
}

const powers = [];
for (const base of ["0", "1", "-1", "2", "-2", "3", "10", "-7", "123456789"]) {
  for (const exponent of ["0", "1", "2", "3", "-1", "-2", "10", "31", "64", "100"]) {
    powers.push({ base, exponent, result: attempt(() => ffi.pow(big(base))(big(exponent))) });
  }
}

const primeCandidates = ["0", "1", "-1", "2", "3", "4", "5", "-7", "9", "25", "47", "49", "91", "97", "561", "1105",
  "7919", "1000003", "2147483647", "3215031751", "4294967291", "4294967297", "18446744073709551557",
  "18446744073709551559", "3825123056546413051", "318665857834031151167461", "170141183460469231731687303715884105727",
  "170141183460469231731687303715884105729", "3317044064679887385961981", "1000000000000000000000000000057"];
for (let i = 0; i < 60; i++) primeCandidates.push(String(2 * Math.floor(random() * 5e8) + 1));
const primes = primeCandidates.map((text) => ({ value: text, prime: ffi.prime(big(text)) }));

console.log(JSON.stringify({ fromBase, fromNumber, unary, binary, powers, primes }));
