// Records AVar.js's behaviour for avar_test.go:
//   bun testdata/generate.js <avar>/src/Effect/AVar.js > testdata/scenarios.json
// A scenario is a list of steps over three variables. Registering steps (put, take, read) log what
// their callback receives and then run their nested steps inside that callback, so drains re-enter.
const AVar = await import(process.argv[2]);

let seed = 71;
const random = () => ((seed = (seed * 1103515245 + 12345) % 2147483648) / 2147483648);
const below = (n) => Math.floor(random() * n);

const util = {
  left: (value) => ({ tag: "Left", value }),
  right: (value) => ({ tag: "Right", value }),
  nothing: { tag: "Nothing" },
  just: (value) => ({ tag: "Just", value }),
  killed: (error) => ({ tag: "Killed", value: error }),
  filled: (value) => ({ tag: "Filled", value }),
  empty: { tag: "Empty" },
};

const show = (value) => (value.value === undefined ? value.tag : `${value.tag} ${value.value}`);

let counter = 0;
function step(depth) {
  const variable = below(3);
  const id = counter++;
  const nested = () => (depth < 2 && random() < 0.4 ? Array.from({ length: 1 + below(2) }, () => step(depth + 1)) : []);
  switch (below(10)) {
    case 0: case 1: return { op: "put", variable, id, value: below(100), then: nested() };
    case 2: case 3: return { op: "take", variable, id, then: nested() };
    case 4: return { op: "read", variable, id, then: nested() };
    case 5: return { op: "cancel", target: below(Math.max(1, counter)) };
    case 6: return { op: random() < 0.3 ? "kill" : "tryPut", variable, value: below(100), message: `boom${id}` };
    case 7: return { op: "tryTake", variable };
    case 8: return { op: "tryRead", variable };
    default: return { op: "status", variable };
  }
}

function run(steps) {
  const variables = [AVar.empty(), AVar._newVar(1000)(), AVar.empty()];
  const cancelers = new Map();
  const log = [];
  const perform = (s) => {
    const callback = (value) => () => {
      log.push(`${s.id} ${show(value)}`);
      s.then.forEach(perform);
    };
    const avar = variables[s.variable];
    switch (s.op) {
      case "put": cancelers.set(s.id, AVar._putVar(util, s.value, avar, callback)()); break;
      case "take": cancelers.set(s.id, AVar._takeVar(util, avar, callback)()); break;
      case "read": cancelers.set(s.id, AVar._readVar(util, avar, callback)()); break;
      case "cancel": { const cancel = cancelers.get(s.target); if (cancel) cancel(); log.push(`cancel ${s.target}`); break; }
      case "kill": AVar._killVar(util, s.message, avar)(); log.push(`kill ${s.variable}`); break;
      case "tryPut": log.push(`tryPut ${s.variable} ${AVar._tryPutVar(util, s.value, avar)()}`); break;
      case "tryTake": log.push(`tryTake ${s.variable} ${show(AVar._tryTakeVar(util, avar)())}`); break;
      case "tryRead": log.push(`tryRead ${s.variable} ${show(AVar._tryReadVar(util, avar)())}`); break;
      case "status": log.push(`status ${s.variable} ${show(AVar._status(util, avar)())}`); break;
    }
  };
  steps.forEach(perform);
  return log;
}

const scenarios = [];
for (let i = 0; i < 300; i++) {
  counter = 0;
  const steps = Array.from({ length: 1 + below(14) }, () => step(0));
  scenarios.push({ steps, log: run(steps) });
}
console.log(JSON.stringify(scenarios));
