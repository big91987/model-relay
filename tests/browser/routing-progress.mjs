const operations = new Set(['panel', 'alias', 'key', 'message', 'stream', 'send', 'complete']);

export function routingProgress(width, sequence, write = console.log, slowAfter = 5000) {
  const started = performance.now();
  const durations = {};
  write(JSON.stringify({routing_call: 'start', utc: new Date().toISOString(), width, sequence}));
  return {
    async step(operation, action) {
      if (!operations.has(operation)) throw new Error('Unknown routing diagnostic operation');
      const began = performance.now();
      const timer = setTimeout(() => write(JSON.stringify({routing_call: 'pending', utc: new Date().toISOString(), width, sequence, operation, elapsed_ms: Math.round(performance.now() - began)})), slowAfter);
      try {
        return await action();
      } finally {
        clearTimeout(timer);
        durations[operation] = Math.round(performance.now() - began);
      }
    },
    finish(ok) {
      write(JSON.stringify({routing_call: 'end', utc: new Date().toISOString(), width, sequence, ok, elapsed_ms: Math.round(performance.now() - started), durations}));
    }
  };
}
