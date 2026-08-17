<script setup lang="ts">
/**
 * AttachmentUploader — 附件上传组件（拖拽/点击 + 图片预览 + 删除）。
 *
 * 工作流：获取预签名 URL → 浏览器直传 MinIO（PUT）→ 刷新附件列表。
 *
 * Props:
 *   workspaceId / projectId — 路由参数
 *   entityType / entityId   — 附件归属实体
 *
 * Events:
 *   @change — 附件列表发生变化（上传成功/删除后触发）
 */
import { onMounted, ref } from "vue";
import { attachmentApi, type Attachment } from "@/api/services/attachment";
import { toast } from "@/lib/toast";

const props = defineProps<{
  workspaceId: number;
  projectId: number;
  entityType: string;
  entityId: number;
}>();

const emit = defineEmits<{ change: [] }>();

const attachments = ref<Attachment[]>([]);
const loading = ref(true);
const error = ref("");
const uploading = ref(false);
const uploadProgress = ref(0); // 0-100
const dragging = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);

const MAX_SIZE = 20 * 1024 * 1024; // 20MB

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function isImage(att: Attachment): boolean {
  return att.content_type.startsWith("image/");
}

async function loadAttachments() {
  if (!props.workspaceId || !props.entityId) return;
  loading.value = true;
  error.value = "";
  try {
    const res = await attachmentApi.listAttachments(
      props.workspaceId, props.projectId, props.entityType, props.entityId,
    );
    attachments.value = res.results;
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "加载附件失败";
  } finally {
    loading.value = false;
  }
}

async function handleFileSelect(files: FileList | null) {
  if (!files || files.length === 0) return;
  const file = files[0];

  if (file.size > MAX_SIZE) {
    error.value = `文件超过 20MB 限制：${file.name}`;
    return;
  }

  uploading.value = true;
  uploadProgress.value = 0;
  error.value = "";
  let storageKey = "";
  try {
    // Step 1: 获取预签名上传 URL
    const presigned = await attachmentApi.getPresignedUploadURL(
      props.workspaceId, props.projectId,
      {
        file_name: file.name,
        content_type: file.type || "application/octet-stream",
        entity_type: props.entityType,
        entity_id: props.entityId,
      },
    );
    storageKey = presigned.storage_key;

    // Step 2: 直接 PUT 到 MinIO（预签名 URL），带进度追踪
    const uploadOk = await putWithProgress(presigned.upload_url, file);
    if (!uploadOk) {
      throw new Error("上传失败");
    }

    // Step 3: 上传确认（服务端 stat 对象 + 写 DB）
    await attachmentApi.confirmUpload(props.workspaceId, props.projectId, {
      file_name: file.name,
      content_type: file.type || "application/octet-stream",
      file_size: file.size,
      entity_type: props.entityType,
      entity_id: props.entityId,
      storage_key: storageKey,
    });

    uploadProgress.value = 100;
    // Step 4: 刷新附件列表
    await loadAttachments();
    toast.success(`已上传 ${file.name}`);
    emit("change");
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "上传失败";
    toast.error(error.value);
  } finally {
    uploading.value = false;
    uploadProgress.value = 0;
    if (fileInput.value) fileInput.value.value = "";
  }
}

/** PUT 文件到预签名 URL，返回是否成功。通过 XMLHttpRequest 获取进度。 */
function putWithProgress(url: string, file: File): Promise<boolean> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", url);
    xhr.setRequestHeader("Content-Type", file.type || "application/octet-stream");
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        uploadProgress.value = Math.round((e.loaded / e.total) * 100);
      }
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve(true);
      } else {
        reject(new Error(`上传失败 (${xhr.status})`));
      }
    };
    xhr.onerror = () => reject(new Error("网络错误，上传中断"));
    xhr.send(file);
  });
}

async function handleDelete(att: Attachment) {
  if (!confirm(`确定删除附件「${att.file_name}」？`)) return;
  try {
    // 优先使用附件自身的 entity_type，回退到 props.entityType（兼容 "issue" 别名）
    const entityType = att.entity_type || props.entityType;
    await attachmentApi.deleteAttachment(props.workspaceId, props.projectId, att.id, entityType);
    attachments.value = attachments.value.filter((a) => a.id !== att.id);
    toast.success(`已删除 ${att.file_name}`);
    emit("change");
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : "删除失败";
    toast.error(error.value);
  }
}

function onDragOver(e: DragEvent) {
  e.preventDefault();
  dragging.value = true;
}

function onDragLeave() {
  dragging.value = false;
}

function onDrop(e: DragEvent) {
  e.preventDefault();
  dragging.value = false;
  handleFileSelect(e.dataTransfer?.files ?? null);
}

function openPicker() {
  fileInput.value?.click();
}

onMounted(loadAttachments);
</script>

<template>
  <div class="attachment-uploader">
    <!-- 上传区 -->
    <div
      class="attachment-uploader__dropzone"
      :class="{ 'attachment-uploader__dropzone--dragging': dragging }"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
      @drop="onDrop"
      @click="openPicker"
    >
      <input
        ref="fileInput"
        type="file"
        class="attachment-uploader__input"
        @change="handleFileSelect(($event.target as HTMLInputElement).files)"
      />
      <template v-if="uploading">
        <span class="attachment-uploader__text">上传中 {{ uploadProgress }}%</span>
        <div class="attachment-uploader__progress">
          <div class="attachment-uploader__progress-bar" :style="{ width: uploadProgress + '%' }"></div>
        </div>
      </template>
      <span v-else class="attachment-uploader__text">📎 点击或拖拽上传附件（≤20MB）</span>
    </div>

    <div v-if="error" class="attachment-uploader__error">{{ error }}</div>

    <!-- 附件列表 -->
    <div v-if="loading" class="attachment-uploader__muted">加载中...</div>
    <div v-else-if="attachments.length === 0" class="attachment-uploader__muted">
      暂无附件
    </div>
    <div v-else class="attachment-uploader__list">
      <div v-for="att in attachments" :key="att.id" class="attachment-uploader__item">
        <!-- 图片附件预览 -->
        <a
          v-if="isImage(att)"
          :href="att.storage_url"
          target="_blank"
          rel="noopener"
          class="attachment-uploader__preview"
        >
          <img :src="att.storage_url" :alt="att.file_name" />
        </a>
        <span v-else class="attachment-uploader__icon">📄</span>

        <div class="attachment-uploader__meta">
          <a
            :href="att.storage_url"
            target="_blank"
            rel="noopener"
            class="attachment-uploader__name"
            :title="att.file_name"
          >
            {{ att.file_name }}
          </a>
          <span class="attachment-uploader__size">{{ formatSize(att.file_size) }}</span>
        </div>

        <button
          class="attachment-uploader__delete"
          title="删除附件"
          @click="handleDelete(att)"
        >
          ✕
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.attachment-uploader {
  margin-top: 8px;
}

.attachment-uploader__dropzone {
  border: 1px dashed var(--border-strong, #d1d5db);
  border-radius: var(--radius-md, 8px);
  padding: 14px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.attachment-uploader__dropzone:hover,
.attachment-uploader__dropzone--dragging {
  border-color: var(--brand-500, #3b82f6);
  background: var(--brand-50, rgba(59,130,246,0.04));
}

.attachment-uploader__input {
  display: none;
}

.attachment-uploader__text {
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.attachment-uploader__error {
  margin-top: 8px;
  font-size: 12px;
  color: var(--danger-500, #ef4444);
}

.attachment-uploader__progress {
  margin-top: 6px;
  height: 4px;
  border-radius: 2px;
  background: var(--border-subtle, #e5e7eb);
  overflow: hidden;
}

.attachment-uploader__progress-bar {
  height: 100%;
  background: var(--brand-500, #3b82f6);
  transition: width 0.2s ease;
}

.attachment-uploader__muted {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
}

.attachment-uploader__list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.attachment-uploader__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border: 1px solid var(--border-subtle, #e5e7eb);
  border-radius: var(--radius-sm, 6px);
  background: var(--surface-2, #f9fafb);
}

.attachment-uploader__preview {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-sm, 4px);
  overflow: hidden;
  flex-shrink: 0;
}

.attachment-uploader__preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.attachment-uploader__icon {
  font-size: 22px;
  flex-shrink: 0;
}

.attachment-uploader__meta {
  flex: 1;
  min-width: 0;
}

.attachment-uploader__name {
  display: block;
  font-size: 13px;
  color: var(--text-primary, #1f2937);
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.attachment-uploader__name:hover {
  color: var(--brand-500, #3b82f6);
}

.attachment-uploader__size {
  font-size: 11px;
  color: var(--text-tertiary, #9ca3af);
}

.attachment-uploader__delete {
  background: none;
  border: none;
  padding: 4px 6px;
  font-size: 12px;
  color: var(--text-tertiary, #9ca3af);
  cursor: pointer;
  flex-shrink: 0;
}

.attachment-uploader__delete:hover {
  color: var(--danger-500, #ef4444);
}
</style>
