<script setup lang="ts">
import SiteShell from '../components/SiteShell.vue'
import { apiDocGroups, eventSourceExample, fetchExample, sseEvents } from '../lib/api-docs'
</script>

<template>
  <SiteShell>
    <section class="api-docs">
      <article class="panel api-docs__hero">
        <p class="eyebrow">Developer reference</p>
        <h1>API 接口文档</h1>
        <p class="panel__lead">
          这是一份面向开发者的公开接口参考页，覆盖当前站点的匿名会话、图片生成、任务查询、
          SSE 事件流和公开画廊接口。
        </p>
      </article>

      <section class="api-docs__intro-grid">
        <article class="panel api-docs__intro-card">
          <p class="eyebrow">Base URL</p>
          <h2>当前站点域名</h2>
          <p class="panel__lead">所有路径都基于当前站点，例如 <code>/api/v1/bootstrap</code>。</p>
        </article>

        <article class="panel api-docs__intro-card">
          <p class="eyebrow">Session</p>
          <h2>匿名会话通过 cookie 保持</h2>
          <p class="panel__lead">
            调用接口时建议带上 <code>credentials: 'include'</code>，服务端会自动创建或复用匿名会话。
          </p>
        </article>

        <article class="panel api-docs__intro-card">
          <p class="eyebrow">Response</p>
          <h2>统一响应结构</h2>
          <p class="panel__lead">
            普通接口统一返回 <code>data</code> 与 <code>error</code> 字段，分页接口包含
            <code>page</code> 与 <code>limit</code>。
          </p>
        </article>
      </section>

      <section
        v-for="group in apiDocGroups"
        :key="group.title"
        class="api-docs__group"
      >
        <div class="section-head api-docs__group-head">
          <div>
            <p class="eyebrow">{{ group.title }}</p>
            <h2>{{ group.title }}</h2>
          </div>
          <p class="section-head__caption">{{ group.description }}</p>
        </div>

        <article
          v-for="item in group.items"
          :key="`${item.method}-${item.path}`"
          class="panel api-endpoint-card"
        >
          <div class="api-endpoint-card__head">
            <div class="api-endpoint-card__title">
              <span class="api-method-badge" :data-method="item.method">{{ item.method }}</span>
              <div>
                <h3>{{ item.title }}</h3>
                <p>{{ item.description }}</p>
              </div>
            </div>
            <code class="api-endpoint-card__path">{{ item.path }}</code>
          </div>

          <div class="api-endpoint-card__body">
            <section v-if="item.request" class="api-code-block">
              <p class="eyebrow">Request</p>
              <pre><code>{{ item.request.join('\n') }}</code></pre>
            </section>

            <section class="api-code-block">
              <p class="eyebrow">Response</p>
              <pre><code>{{ item.response.join('\n') }}</code></pre>
            </section>
          </div>
        </article>
      </section>

      <section class="api-docs__group">
        <div class="section-head api-docs__group-head">
          <div>
            <p class="eyebrow">SSE stream</p>
            <h2>任务事件流</h2>
          </div>
          <p class="section-head__caption">
            使用 <code>GET /api/v1/generations/:taskId/stream</code> 订阅任务状态变化。
          </p>
        </div>

        <article class="panel api-sse-card">
          <div class="api-sse-card__events">
            <article
              v-for="event in sseEvents"
              :key="event.type"
              class="api-sse-card__event"
            >
              <strong>{{ event.type }}</strong>
              <p>{{ event.description }}</p>
            </article>
          </div>
        </article>
      </section>

      <section class="api-docs__group">
        <div class="section-head api-docs__group-head">
          <div>
            <p class="eyebrow">Examples</p>
            <h2>调用示例</h2>
          </div>
          <p class="section-head__caption">下面两段示例分别用于任务提交和 EventSource 订阅。</p>
        </div>

        <div class="api-docs__examples">
          <article class="panel api-code-block">
            <p class="eyebrow">fetch</p>
            <pre><code>{{ fetchExample.join('\n') }}</code></pre>
          </article>

          <article class="panel api-code-block">
            <p class="eyebrow">EventSource</p>
            <pre><code>{{ eventSourceExample.join('\n') }}</code></pre>
          </article>
        </div>
      </section>
    </section>
  </SiteShell>
</template>
