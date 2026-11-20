// Records JavaScript's behaviour for argonaut_test.go:
//   bun testdata/generate.js <argonaut-core>/src/Data/Argonaut/Core.js > testdata/cases.json
// Each parse case is JSON.parse's result (re-stringified with several indents) or its message;
// each compare case is argonaut-core's own _compare on two parsed values.
const core = await import(process.argv[2]);

let seed = 20261115;
const random = () => ((seed = (seed * 1103515245 + 12345) % 2147483648) / 2147483648);
const pick = (items) => items[Math.floor(random() * items.length)];

const handwritten = [
  "", " ", "{", "}", "[", "]", "[1,]", "[,1]", "{\"a\":1,}", "{\"a\" 1}", "{a:1}", "{'a':1}",
  "\"abc", "\"a\\x\"", "\"a\\u12\"", "\"a\nb\"", "\"\t\"", "tru", "true ", "truex", "nul", "nulll",
  "falsey", "01", "-", "-0", "1.", ".5", "1e", "1e+", "1E5", "+1", "0x10", "NaN", "Infinity",
  "[1 2]", "{\"a\":1 \"b\":2}", "1 2", "\u00a0 1", "\ufeff1", "[1]x", "{\"a\":}", "{\"a\"",
  "{\"a\":1", "[1", "[\"a\"", "\"\\uD83D\\uDE00\"", "1e400", "-1e400", "1e-400",
  "123456789012345678901234567890", "{\"__proto__\":1}", "{\"b\":1,\"a\":2,\"b\":3}",
  "{\"2\":1,\"1\":2,\"x\":3,\"01\":4,\"4294967295\":5,\"4294967294\":6,\"-1\":7,\"1.5\":8}",
  "\"\\/\"", "[\n1\r\n,\t2]", "\u000b1", "/*c*/1", "[-]", "-a", "1.5e-7", "5e-324", "1e21",
  "1e20", "0.1e1", "-0.0", "\"\\u0000\"", "\"\u0001\"", "{\"\":1}", "[[[[]]]]", "@", "\"\\",
  "[true,false,null]", "{\"a\":[1,{\"b\":null}]}", "[1 @", "(", ")", ";", "=", ":", "{\"a\":1 @",
  "\"\\é\"", "\"\\\u0001\"", "é", "_x", "$", "[1,@]", "{\"a\":1,\"b\"}", "{\"a\":1,", "[1,",
  "{\"a\":", "{,}", "[\"\\u00zz\"]", "{1:2}", "-01", "-.5", "1.5.5", "1e5e5", "[]]", "{}}", "{} ",
  "\"\\u", "\"\\u1", "  true\n", "[1,2,3,]", "{\"a\":1}x", "t", "f", "n", "fals", "nulx", "trux",
  "tRUE", "{\"a\"x", "{\"a\":1:", "[1:2]", "\"a\u2028b\"", "\u2028", "123.456e+2",
  "0.30000000000000004", "9007199254740993", "-0e0", "1E-2", "\"\\u00e9\\u00E9\"",
  "\"tab\\there \\\"quoted\\\" back\\\\slash \\b\\f\\n\\r\"", "\"\\u001f\\u007f\\u2028\"",
  "{\"nested\":{\"deeper\":{\"deepest\":[1,[2,[3,{}]],[]]}},\"empty\":{},\"list\":[]}",
  "[0.000001,0.0000001,1e-7,123e-20,1e300,4.35,0.1,100,1e100,-5e-324]", "\"\\uD83D\"",
  "\"\\uDE00\\uD83D\"", "\"x\\uD83Dy\"", "\"\\uD83D\\u0041\"", "\"😀\"", "{\"😀\":1,\"é\":2,\"\":3}",
];

const fragments = ["{", "}", "[", "]", ",", ":", "\"", "\\", "\\u", "\\n", "0", "1", "9", "-", "+",
  ".", "e", "E", "true", "false", "null", "tr", "x", "'", " ", "\n", "\t", "é", "@", "\"a\"",
  "\"b\":", "12", "3.5", "1e3", "\\uD83D", "00", "$"];

function randomJson(depth) {
  const choice = Math.floor(random() * (depth > 3 ? 4 : 6));
  switch (choice) {
    case 0: return null;
    case 1: return random() < 0.5;
    case 2: return pick([0, -1, 1.5, 1e21, 1e-7, 123456789, -0.25, 2 ** 53, 0.1 + 0.2]);
    case 3: return pick(["", "a", "é", "😀", "\"", "\\", "\n", "\u0001", "x y", "1"]);
    case 4: return Array.from({ length: Math.floor(random() * 4) }, () => randomJson(depth + 1));
    default: {
      const value = {};
      for (let i = Math.floor(random() * 4); i > 0; i--) {
        value[pick(["a", "b", "1", "0", "10", "z", "", "é", "01"])] = randomJson(depth + 1);
      }
      return value;
    }
  }
}

const inputs = [...handwritten];
for (let i = 0; i < 400; i++) {
  inputs.push(Array.from({ length: 1 + Math.floor(random() * 8) }, () => pick(fragments)).join(""));
}
for (let i = 0; i < 200; i++) {
  const text = JSON.stringify(randomJson(0));
  const at = Math.floor(random() * (text.length + 1));
  inputs.push(random() < 0.5 ? text : text.slice(0, at) + pick(fragments) + text.slice(at));
}

const parse = inputs.map((input) => {
  try {
    const value = JSON.parse(input);
    return { input, compact: JSON.stringify(value), indented: [-1, 1, 2, 11].map((indent) => JSON.stringify(value, null, indent)) };
  } catch (error) {
    return { input, message: error.message };
  }
});

const values = ["null", "true", "false", "0", "-1", "1.5", "\"\"", "\"a\"", "\"b\"", "\"é\"", "\"😀\"", "\"\uffff\"",
  "[]", "[1]", "[1,2]", "[2]", "[null]", "{}", "{\"a\":1}", "{\"a\":2}", "{\"b\":1}", "{\"a\":1,\"b\":2}",
  "{\"b\":2,\"a\":1}", "{\"a\":[1,{\"c\":null}]}", "{\"a\":[1,{\"c\":false}]}"];
const ordering = { EQ: "EQ", GT: "GT", LT: "LT" };
const compare = [];
for (const left of values) {
  for (const right of values) {
    compare.push({ left, right, order: core._compare(ordering.EQ, ordering.GT, ordering.LT, JSON.parse(left), JSON.parse(right)) });
  }
}
compare.push({ left: "NaN", right: "NaN", order: core._compare("EQ", "GT", "LT", NaN, NaN) });
compare.push({ left: "NaN", right: "1", order: core._compare("EQ", "GT", "LT", NaN, 1) });
compare.push({ left: "1", right: "NaN", order: core._compare("EQ", "GT", "LT", 1, NaN) });

const numbers = [NaN, Infinity, -Infinity, -0, 0, 1e21, 1e-7, 5e-324, 0.1, 123456789012].map((number) => ({
  number: Object.is(number, -0) ? "-0" : String(number),
  stringified: core.stringify(number),
}));

console.log(JSON.stringify({ parse, compare, numbers }, null, 1));
