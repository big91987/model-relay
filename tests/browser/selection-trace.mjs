export function observeSelection(node) {
  const document = node.ownerDocument;
  const originalOptions = [...node.options];
  const events = [];
  const boundaries = [];
  let eventCount = 0;
  let mutationCount = 0;
  let stopped = false;
  const snapshot = () => ({selectedIndex: node.selectedIndex, optionCount: node.options.length, focused: node === document.activeElement, documentFocused: document.hasFocus(), disabled: node.disabled, connected: node.isConnected, sameOptions: node.options.length === originalOptions.length && originalOptions.every((option, index) => option === node.options[index]), mutationCount});
  const event = observed => {
    const printable = typeof observed.key === 'string' && /^[\x20-\x7e]$/.test(observed.key);
    if (['keydown', 'keyup'].includes(observed.type) && !['Home', 'ArrowDown', 'Tab'].includes(observed.key) && !printable) return;
    if (!['keydown', 'keyup'].includes(observed.type) && observed.target !== node) return;
    eventCount++;
    events.push({type: observed.type, key: ['Home', 'ArrowDown', 'Tab'].includes(observed.key) ? observed.key : printable ? 'printable' : 'none', target: observed.target === node ? 'original-select' : observed.target?.tagName === 'SELECT' ? 'other-select' : 'other', prevented: observed.defaultPrevented, trusted: observed.isTrusted, ...snapshot()});
    if (events.length > 24) events.splice(4, 1);
  };
  const types = ['keydown', 'keyup', 'input', 'change', 'focusin', 'focusout'];
  for (const type of types) document.addEventListener(type, event);
  const observer = new document.defaultView.MutationObserver(records => { mutationCount += records.length; });
  observer.observe(node.parentNode, {childList: true, subtree: true, attributes: true});
  const result = () => ({events: [...events], boundaries: [...boundaries], eventCount, omittedEvents: Math.max(0, eventCount - events.length), mutationCount});
  return {
    sample(boundary) { const state = {boundary, ...snapshot()}; boundaries.push(state); if (boundaries.length > 12) boundaries.shift(); return {...state, ...result()}; },
    stop() { if (!stopped) { stopped = true; observer.disconnect(); for (const type of types) document.removeEventListener(type, event); } return result(); }
  };
}
