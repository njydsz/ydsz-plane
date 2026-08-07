// 此文件为历史遗留占位，已被迁移替代。
//
// 性能压测脚本已统一迁移至 tests/perf/：
//   - tests/perf/smoke-test.js   冒烟测试（1 VU / 1 迭代）
//   - tests/perf/load-test.js    负载测试（10→100 VU，P95<200ms 断言）
//   - tests/perf/stress-test.js  压力测试（200 VU 恒定 3 分钟）
//
// 运行方式（等价 Makefile 目标）：
//   make perf-smoke && make perf-load && make perf-stress
//
// 本文件不再被 CI / Makefile / 文档引用，仅作迁移记录保留。
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:8080/api/v1';

export const options = {
  vus: 1,
  iterations: 1,
};

export default function () {
  const health = http.get(`${BASE_URL.replace('/api/v1', '')}/healthz`);
  check(health, { 'healthz: 200': (r) => r.status === 200 });
}
