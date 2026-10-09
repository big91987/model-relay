export function cooldownObservation(value, alias, now = Date.now()) {
  const models = Array.isArray(value?.data) ? value.data : [];
  const matches = models.filter(model => model?.alias === alias);
  const candidates = matches.flatMap(model => Array.isArray(model.candidates) ? model.candidates : []);
  const remaining = candidates.map(candidate => typeof candidate.cooldown_until === 'string' ? Date.parse(candidate.cooldown_until) - now : NaN).filter(Number.isFinite);
  return {
    model_count: models.length,
    matching_models: matches.length,
    has_next_page: typeof value?.next_cursor === 'string' && value.next_cursor.length > 0,
    candidate_count: candidates.length,
    cooling_count: candidates.filter(candidate => candidate.generation_health === '暂时冷却').length,
    expired_count: candidates.filter(candidate => candidate.generation_health === '待重新尝试').length,
    failed_count: candidates.filter(candidate => candidate.generation_health === '最近生成失败（仍参与）').length,
    minimum_remaining_ms: remaining.length ? Math.round(Math.min(...remaining)) : null
  };
}
