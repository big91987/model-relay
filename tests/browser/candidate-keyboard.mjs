import assert from 'node:assert/strict';

export async function typeAheadCandidate(control, keyboard, label, sample) {
  const prefix = label.split(' ')[0];
  assert.match(prefix, /^[A-Za-z][A-Za-z0-9-]*$/, 'candidate fixture needs a printable ASCII prefix');
  const valid = await control.evaluate((node, target) => {
    const options = [...node.options];
    const matching = options.filter(option => option.label.toLowerCase().startsWith(target.prefix.toLowerCase()));
    return !node.disabled && matching.length === 1 && !matching[0].disabled && matching[0].label === target.label;
  }, {prefix, label});
  assert.equal(valid, true, 'candidate prefix must identify one enabled target');
  await sample('before type-ahead');
  await keyboard.type(prefix);
  await sample('after type-ahead');
  await keyboard.press('Tab');
  await sample('after Tab');
}
