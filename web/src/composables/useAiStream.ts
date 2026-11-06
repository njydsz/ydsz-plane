/**
 * useAiStream — AI 流式生成（SSE）。
 *
 * 能力：
 *   - 通过 fetch + ReadableStream 消费后端 SSE；
 *   - 逐 token 回调（onDelta），支持实时渲染；
 *   - 自动处理完成/错误事件；
 *   - 支持中止（AbortController）；
 *   - 与现有 axios 实例的 baseURL 保持一致。
 *
 * 用法：
 *   const { start, abort, streaming, result, error } = useAiStream()
 *   start({ task: 'assist', context: '...', full_text: '...' })
 */
import { ref } from "vue"

import { getAccessToken } from "@/api/client"

export interface AiStreamRequest {
  task: "assist" | "rewrite" | "grammar" | "chat"
  context?: string
  full_text?: string
  style?: string
  language?: string
  max_tokens?: number
  messages?: Array<{ role: string; content: string }>
}

export interface AiStreamOptions {
  /** 增量文本回调 */
  onDelta?: (text: string) => void
  /** 完成回调 */
  onDone?: (fullText: string, usage?: { prompt_tokens: number; completion_tokens: number }) => void
  /** 错误回调 */
  onError?: (message: string) => void
}

export function useAiStream() {
  const streaming = ref(false)
  const result = ref("")
  const error = ref("")
  const usage = ref<{ prompt_tokens: number; completion_tokens: number } | null>(null)

  let controller: AbortController | null = null

  /**
   * 启动流式请求。
   */
  async function start(req: AiStreamRequest, opts: AiStreamOptions = {}): Promise<string> {
    abort()
    controller = new AbortController()

    result.value = ""
    error.value = ""
    usage.value = null
    streaming.value = true

    try {
      const token = getAccessToken()

      const resp = await fetch("/api/v1/ai/stream", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "text/event-stream",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(req),
        signal: controller.signal,
      })

      if (!resp.ok) {
        const text = await resp.text()
        throw new Error(`HTTP ${resp.status}: ${text}`)
      }

      const reader = resp.body?.getReader()
      if (!reader) {
        throw new Error("响应流不可用")
      }

      const decoder = new TextDecoder()
      let buffer = ""
      let fullText = ""

      for (;;) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // 按双换行切分 SSE 事件
        const events = buffer.split("\n\n")
        buffer = events.pop() ?? ""

        for (const ev of events) {
          const parsed = parseSSEEvent(ev)
          if (!parsed) continue

          switch (parsed.event) {
            case "delta":
              fullText += parsed.data.text ?? ""
              result.value = fullText
              opts.onDelta?.(parsed.data.text ?? "")
              break
            case "done":
              usage.value = parsed.data.usage ?? null
              streaming.value = false
              opts.onDone?.(fullText, parsed.data.usage)
              return fullText
            case "error":
              throw new Error(parsed.data.message ?? "生成失败")
          }
        }
      }

      streaming.value = false
      return fullText
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : "生成失败"
      error.value = message
      streaming.value = false
      opts.onError?.(message)
      return ""
    }
  }

  /** 中止当前请求 */
  function abort() {
    controller?.abort()
    controller = null
    streaming.value = false
  }

  return { streaming, result, error, usage, start, abort }
}

// ----------------------------------------------------------------------
// SSE 解析
// ----------------------------------------------------------------------

interface SSEData {
  text?: string
  message?: string
  finish_reason?: string
  usage?: { prompt_tokens: number; completion_tokens: number }
}

function parseSSEEvent(raw: string): { event: string; data: SSEData } | null {
  let event = "message"
  let dataLine = ""

  for (const line of raw.split("\n")) {
    if (line.startsWith("event:")) {
      event = line.slice(6).trim()
    } else if (line.startsWith("data:")) {
      dataLine = line.slice(5).trim()
    }
  }

  if (!dataLine) return null

  try {
    return { event, data: JSON.parse(dataLine) as SSEData }
  } catch {
    return null
  }
}
