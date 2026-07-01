<template>
  <div class="space-y-6">
    <section class="card">
      <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ localText('帮助中心', 'Help Center') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ localText('通过表单维护教程、FAQ 和附件；保存只更新草稿，发布后用户侧才会看到。', 'Manage tutorials, FAQs, and attachments with forms. Save updates the draft only; publish makes it visible to users.') }}
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <button type="button" class="btn btn-secondary" :disabled="busy" @click="loadConfig">
              <Icon name="refresh" size="sm" />
              {{ localText('刷新', 'Refresh') }}
            </button>
            <button type="button" class="btn btn-secondary" :disabled="busy" @click="saveDraftOnly">
              <Icon name="document" size="sm" />
              {{ saving ? localText('保存中...', 'Saving...') : localText('保存草稿', 'Save Draft') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="busy" @click="publishCurrentDraft">
              <Icon name="upload" size="sm" />
              {{ publishing ? localText('发布中...', 'Publishing...') : localText('发布', 'Publish') }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="loading || !draft" class="flex min-h-[360px] items-center justify-center p-8">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <div v-else class="grid grid-cols-1 gap-6 p-6 2xl:grid-cols-[minmax(0,1fr)_360px]">
        <div class="space-y-6">
          <section class="rounded-lg border border-gray-200 p-5 dark:border-dark-700">
            <div class="flex items-center justify-between gap-4">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ localText('基础设置', 'Basics') }}</h3>
              <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                <input v-model="draft.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                {{ localText('启用帮助中心', 'Enabled') }}
              </label>
            </div>
            <div class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
              <label class="block">
                <span class="input-label">{{ localText('标题', 'Title') }}</span>
                <input v-model="draft.title" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">Base URL</span>
                <input v-model="draft.base_url" class="input mt-1" placeholder="https://useaifor.me" />
              </label>
              <label class="block lg:col-span-2">
                <span class="input-label">{{ localText('描述', 'Description') }}</span>
                <textarea v-model="draft.description" class="input mt-1 min-h-[84px]"></textarea>
              </label>
            </div>
          </section>

          <section class="rounded-lg border border-gray-200 p-5 dark:border-dark-700">
            <div class="flex items-center justify-between gap-4">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ localText('创建 Key 后弹窗', 'Key Created Prompt') }}</h3>
              <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                <input v-model="draft.key_created_prompt.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                {{ localText('启用弹窗', 'Enabled') }}
              </label>
            </div>
            <div class="mt-4 grid grid-cols-1 gap-4 lg:grid-cols-2">
              <label class="block">
                <span class="input-label">{{ localText('弹窗标题', 'Title') }}</span>
                <input v-model="draft.key_created_prompt.title" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">{{ localText('不再提示文案', 'Dismiss Label') }}</span>
                <input v-model="draft.key_created_prompt.dismiss_label" class="input mt-1" />
              </label>
              <label class="block lg:col-span-2">
                <span class="input-label">{{ localText('弹窗说明', 'Description') }}</span>
                <textarea v-model="draft.key_created_prompt.description" class="input mt-1 min-h-[84px]"></textarea>
              </label>
              <label class="block">
                <span class="input-label">{{ localText('主按钮文案', 'Primary Label') }}</span>
                <input v-model="draft.key_created_prompt.primary_action_label" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">{{ localText('主按钮链接', 'Primary URL') }}</span>
                <input v-model="draft.key_created_prompt.primary_action_url" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">{{ localText('次按钮文案', 'Secondary Label') }}</span>
                <input v-model="draft.key_created_prompt.secondary_action_label" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">{{ localText('次按钮链接', 'Secondary URL') }}</span>
                <input v-model="draft.key_created_prompt.secondary_action_url" class="input mt-1" />
              </label>
            </div>
          </section>

          <section class="rounded-lg border border-gray-200 p-5 dark:border-dark-700">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ localText('教程块', 'Tutorials') }}</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('控制用户侧左侧导航和每个客户端的教程内容。', 'Controls the user-side client navigation and tutorial content.') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" @click="addTutorial">
                <Icon name="plus" size="sm" />
                {{ localText('添加教程', 'Add Tutorial') }}
              </button>
            </div>

            <div class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-[260px_minmax(0,1fr)]">
              <div class="space-y-2">
                <button
                  v-for="(tutorial, index) in draft.tutorials"
                  :key="tutorial.id || index"
                  type="button"
                  class="w-full rounded-md border px-3 py-2 text-left text-sm transition"
                  :class="selectedTutorialIndex === index ? 'border-primary-300 bg-primary-50 text-primary-700 dark:border-primary-700 dark:bg-primary-900/20 dark:text-primary-300' : 'border-gray-200 text-gray-700 hover:bg-gray-50 dark:border-dark-700 dark:text-gray-300 dark:hover:bg-dark-800'"
                  @click="selectedTutorialIndex = index"
                >
                  <span class="flex items-center justify-between gap-2">
                    <span class="truncate font-medium">{{ tutorial.title || localText('未命名教程', 'Untitled') }}</span>
                    <span class="text-xs text-gray-400">#{{ tutorial.sort_order }}</span>
                  </span>
                  <span class="mt-1 block truncate text-xs text-gray-500">{{ tutorial.id }}</span>
                </button>
              </div>

              <div v-if="selectedTutorial" class="space-y-5 rounded-lg bg-gray-50 p-4 dark:bg-dark-900">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="selectedTutorial.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    {{ localText('启用', 'Enabled') }}
                  </label>
                  <div class="flex flex-wrap gap-2">
                    <button type="button" class="btn btn-secondary text-xs" @click="moveTutorial(-1)">{{ localText('上移', 'Up') }}</button>
                    <button type="button" class="btn btn-secondary text-xs" @click="moveTutorial(1)">{{ localText('下移', 'Down') }}</button>
                    <button type="button" class="btn btn-secondary text-xs" @click="duplicateTutorial">
                      <Icon name="copy" size="xs" />
                      {{ localText('复制教程', 'Duplicate Tutorial') }}
                    </button>
                    <button type="button" class="btn btn-danger text-xs" @click="removeTutorial">{{ localText('删除', 'Delete') }}</button>
                  </div>
                </div>

                <div class="grid grid-cols-1 gap-4 lg:grid-cols-4">
                  <label class="block">
                    <span class="input-label">ID</span>
                    <input v-model="selectedTutorial.id" class="input mt-1" />
                  </label>
                  <label class="block">
                    <span class="input-label">{{ localText('排序', 'Sort') }}</span>
                    <input v-model.number="selectedTutorial.sort_order" type="number" class="input mt-1" />
                  </label>
                  <label class="block lg:col-span-2">
                    <span class="input-label">{{ localText('标题', 'Title') }}</span>
                    <input v-model="selectedTutorial.title" class="input mt-1" />
                  </label>
                  <label class="block">
                    <span class="input-label">{{ localText('标签', 'Badge') }}</span>
                    <input v-model="selectedTutorial.badge" class="input mt-1" />
                  </label>
                  <label class="block lg:col-span-3">
                    <span class="input-label">{{ localText('摘要', 'Summary') }}</span>
                    <input v-model="selectedTutorial.summary" class="input mt-1" />
                  </label>
                  <label class="block lg:col-span-4">
                    <span class="input-label">{{ localText('正文 Markdown', 'Markdown Content') }}</span>
                    <textarea v-model="selectedTutorial.content_md" class="input mt-1 min-h-[180px] font-mono text-sm"></textarea>
                  </label>
                </div>

                <EditableList
                  :title="localText('步骤', 'Steps')"
                  :items="selectedTutorial.steps"
                  :add-label="localText('添加步骤', 'Add Step')"
                  @add="selectedTutorial.steps.push(emptyStep())"
                  @remove="selectedTutorial.steps.splice($event, 1)"
                >
                  <template #default="{ item }">
                    <div class="space-y-4">
                      <div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
                        <input v-model="item.title" class="input" :placeholder="localText('步骤标题', 'Step title')" />
                        <textarea v-model="item.description" class="input min-h-[74px] lg:col-span-2" :placeholder="localText('步骤说明', 'Step description')"></textarea>
                      </div>

                      <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
                        <div class="rounded-md border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900">
                          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ localText('步骤图片', 'Step Images') }}</h5>
                            <div class="flex flex-wrap gap-2">
                              <button type="button" class="btn btn-secondary text-xs" @click="openImageDialog(item)">
                                <Icon name="plus" size="xs" />
                                {{ localText('添加图片', 'Add Image') }}
                              </button>
                              <label class="btn btn-secondary cursor-pointer text-xs">
                                <Icon name="upload" size="xs" />
                                {{ uploadingAttachment ? localText('上传中...', 'Uploading...') : localText('上传图片', 'Upload Image') }}
                                <input type="file" accept="image/*" class="sr-only" :disabled="uploadingAttachment" @change="handleStepImageUpload(item, $event)" />
                              </label>
                            </div>
                          </div>
                          <div v-if="item.images?.length" class="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2">
                            <div v-for="(image, imageIndex) in item.images" :key="`${image.url}-${imageIndex}`" class="overflow-hidden rounded-md border border-gray-100 bg-white dark:border-dark-700 dark:bg-dark-800">
                              <div class="flex aspect-video items-center justify-center bg-gray-100 px-3 text-center text-xs text-gray-500 dark:bg-dark-900 dark:text-gray-400">
                                <span class="line-clamp-3 break-all">{{ image.file_name || image.url || localText('未设置图片', 'No image selected') }}</span>
                              </div>
                              <div class="space-y-2 p-3">
                                <input v-model="image.label" class="input" :placeholder="localText('图片名称', 'Image label')" />
                                <input v-model="image.url" class="input text-xs" placeholder="/api/v1/help-center/attachments/image.png" />
                                <div class="grid grid-cols-1 gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
                                  <input v-model="image.file_name" class="input text-xs" :placeholder="localText('文件名', 'File name')" />
                                  <button type="button" class="btn btn-danger text-xs" @click="item.images.splice(imageIndex, 1)">{{ localText('删除', 'Delete') }}</button>
                                </div>
                              </div>
                            </div>
                          </div>
                        </div>

                        <div class="rounded-md border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900">
                          <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                            <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ localText('步骤代码块', 'Step Code Blocks') }}</h5>
                            <button type="button" class="btn btn-secondary text-xs" @click="openCodeBlockDialog(item.code_blocks, item.title || localText('步骤', 'Step'))">
                              <Icon name="plus" size="xs" />
                              {{ localText('添加代码块', 'Add Code Block') }}
                            </button>
                          </div>
                          <div v-if="item.code_blocks?.length" class="mt-3 space-y-2">
                            <div v-for="(block, blockIndex) in item.code_blocks" :key="`${block.title}-${blockIndex}`" class="grid grid-cols-1 gap-2 rounded-md border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800 lg:grid-cols-4">
                              <input v-model="block.title" class="input" :placeholder="localText('标题', 'Title')" />
                              <input v-model="block.language" class="input" placeholder="bash" />
                              <textarea v-model="block.content" class="input min-h-[88px] font-mono text-sm lg:col-span-2" :placeholder="localText('代码内容', 'Code content')"></textarea>
                              <button type="button" class="btn btn-danger text-xs lg:col-start-4" @click="item.code_blocks.splice(blockIndex, 1)">{{ localText('删除', 'Delete') }}</button>
                            </div>
                          </div>
                        </div>
                      </div>

                      <div class="rounded-md border border-gray-100 bg-gray-50 p-3 dark:border-dark-700 dark:bg-dark-900">
                        <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                          <h5 class="text-xs font-semibold text-gray-700 dark:text-gray-300">{{ localText('\u6b65\u9aa4\u9644\u4ef6', 'Step Attachments') }}</h5>
                          <label class="btn btn-secondary cursor-pointer text-xs">
                            <Icon name="upload" size="xs" />
                            {{ uploadingAttachment ? localText('\u4e0a\u4f20\u4e2d...', 'Uploading...') : localText('\u4e0a\u4f20\u9644\u4ef6', 'Upload Attachment') }}
                            <input type="file" class="sr-only" :disabled="uploadingAttachment" @change="handleStepAttachmentUpload(item, $event)" />
                          </label>
                        </div>
                        <div class="mt-3 space-y-2">
                          <div v-for="(attachment, attachmentIndex) in item.attachments" :key="`${attachment.url}-${attachmentIndex}`" class="grid grid-cols-1 gap-2 rounded-md border border-gray-100 bg-white p-3 dark:border-dark-700 dark:bg-dark-800 lg:grid-cols-[1fr_1.6fr_1fr_auto]">
                            <input v-model="attachment.label" class="input" :placeholder="localText('\u9644\u4ef6\u540d\u79f0', 'Attachment label')" />
                            <input v-model="attachment.url" class="input text-xs" placeholder="/api/v1/help-center/attachments/file.pdf" />
                            <input v-model="attachment.file_name" class="input text-xs" :placeholder="localText('\u6587\u4ef6\u540d', 'File name')" />
                            <button type="button" class="btn btn-danger text-xs" @click="item.attachments.splice(attachmentIndex, 1)">{{ localText('\u5220\u9664', 'Delete') }}</button>
                          </div>
                          <button type="button" class="btn btn-secondary text-xs" @click="item.attachments.push({ label: '', url: '', file_name: '' })">
                            <Icon name="plus" size="xs" />
                            {{ localText('\u624b\u52a8\u6dfb\u52a0\u9644\u4ef6', 'Add Manually') }}
                          </button>
                        </div>
                      </div>
                    </div>
                  </template>
                </EditableList>

                <EditableList
                  :title="localText('代码块', 'Code Blocks')"
                  :items="selectedTutorial.code_blocks"
                  :add-label="localText('添加代码块', 'Add Code Block')"
                  @add="openCodeBlockDialog(selectedTutorial.code_blocks, selectedTutorial.title || localText('教程', 'Tutorial'))"
                  @remove="selectedTutorial.code_blocks.splice($event, 1)"
                >
                  <template #default="{ item }">
                    <div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
                      <input v-model="item.title" class="input" :placeholder="localText('标题', 'Title')" />
                      <input v-model="item.language" class="input" placeholder="bash" />
                      <textarea v-model="item.content" class="input min-h-[96px] font-mono text-sm lg:col-span-2" :placeholder="localText('代码内容', 'Code content')"></textarea>
                    </div>
                  </template>
                </EditableList>

                <EditableList
                  :title="localText('链接', 'Links')"
                  :items="selectedTutorial.links"
                  :add-label="localText('添加链接', 'Add Link')"
                  @add="selectedTutorial.links.push({ label: '', url: '' })"
                  @remove="selectedTutorial.links.splice($event, 1)"
                >
                  <template #default="{ item }">
                    <div class="grid grid-cols-1 gap-3 lg:grid-cols-2">
                      <input v-model="item.label" class="input" :placeholder="localText('链接名称', 'Label')" />
                      <input v-model="item.url" class="input" placeholder="/keys" />
                    </div>
                  </template>
                </EditableList>

                <div class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800">
                  <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ localText('附件', 'Attachments') }}</h4>
                    <label class="btn btn-secondary cursor-pointer text-xs">
                      <Icon name="upload" size="xs" />
                      {{ uploadingAttachment ? localText('上传中...', 'Uploading...') : localText('上传附件', 'Upload Attachment') }}
                      <input type="file" class="sr-only" :disabled="uploadingAttachment" @change="handleAttachmentUpload" />
                    </label>
                  </div>
                  <div class="mt-3 space-y-3">
                    <div v-for="(attachment, index) in selectedTutorial.attachments" :key="`${attachment.url}-${index}`" class="grid grid-cols-1 gap-3 rounded-md border border-gray-100 p-3 dark:border-dark-700 lg:grid-cols-[1fr_1.6fr_1fr_auto]">
                      <input v-model="attachment.label" class="input" :placeholder="localText('附件名称', 'Label')" />
                      <input v-model="attachment.url" class="input" placeholder="/api/v1/help-center/attachments/file.pdf" />
                      <input v-model="attachment.file_name" class="input" :placeholder="localText('文件名', 'File name')" />
                      <button type="button" class="btn btn-danger text-xs" @click="selectedTutorial.attachments.splice(index, 1)">{{ localText('删除', 'Delete') }}</button>
                    </div>
                    <button type="button" class="btn btn-secondary text-xs" @click="selectedTutorial.attachments.push({ label: '', url: '', file_name: '' })">
                      <Icon name="plus" size="xs" />
                      {{ localText('手动添加附件', 'Add Manually') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-lg border border-gray-200 p-5 dark:border-dark-700">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">FAQ</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('配置用户侧快捷问答，答案支持 Markdown。', 'Configure user-facing quick answers. Markdown is supported.') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" @click="addFAQ">
                <Icon name="plus" size="sm" />
                {{ localText('添加 FAQ', 'Add FAQ') }}
              </button>
            </div>
            <div class="mt-4 space-y-4">
              <div v-for="(faq, index) in draft.faqs" :key="faq.id || index" class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900">
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <label class="inline-flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                    <input v-model="faq.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                    {{ localText('启用', 'Enabled') }}
                  </label>
                  <div class="flex flex-wrap gap-2">
                    <button type="button" class="btn btn-secondary text-xs" @click="moveFAQ(index, -1)">{{ localText('上移', 'Up') }}</button>
                    <button type="button" class="btn btn-secondary text-xs" @click="moveFAQ(index, 1)">{{ localText('下移', 'Down') }}</button>
                    <button type="button" class="btn btn-danger text-xs" @click="draft.faqs.splice(index, 1)">{{ localText('删除', 'Delete') }}</button>
                  </div>
                </div>
                <div class="mt-3 grid grid-cols-1 gap-3 lg:grid-cols-4">
                  <input v-model="faq.id" class="input" placeholder="faq-id" />
                  <input v-model.number="faq.sort_order" type="number" class="input" :placeholder="localText('排序', 'Sort')" />
                  <input v-model="faq.question" class="input lg:col-span-2" :placeholder="localText('问题', 'Question')" />
                  <input :value="faq.tags.join(', ')" class="input lg:col-span-4" :placeholder="localText('标签，用逗号分隔', 'Tags separated by commas')" @input="updateFAQTags(faq, $event)" />
                  <textarea v-model="faq.answer_md" class="input min-h-[120px] font-mono text-sm lg:col-span-4" :placeholder="localText('答案 Markdown', 'Answer Markdown')"></textarea>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-lg border border-gray-200 p-5 dark:border-dark-700">
            <button type="button" class="flex w-full items-center justify-between text-left" @click="showAdvancedJson = !showAdvancedJson">
              <span>
                <span class="block text-base font-semibold text-gray-900 dark:text-white">{{ localText('高级 JSON', 'Advanced JSON') }}</span>
                <span class="mt-1 block text-sm text-gray-500 dark:text-gray-400">{{ localText('用于批量复制、备份或快速替换配置。', 'Use for bulk copy, backup, or quick replacement.') }}</span>
              </span>
              <Icon :name="showAdvancedJson ? 'chevronUp' : 'chevronDown'" size="sm" />
            </button>
            <div v-if="showAdvancedJson" class="mt-4 space-y-3">
              <textarea v-model="advancedJson" spellcheck="false" class="input min-h-[360px] font-mono text-sm leading-6" :class="{ 'border-red-500 dark:border-red-500': !!jsonError }"></textarea>
              <div class="flex flex-wrap items-center justify-between gap-3">
                <p v-if="jsonError" class="text-sm text-red-500">{{ jsonError }}</p>
                <span v-else class="text-sm text-gray-500">{{ localText('修改后点击应用到表单，再保存草稿。', 'Apply to form before saving the draft.') }}</span>
                <button type="button" class="btn btn-secondary" @click="applyAdvancedJson">{{ localText('应用到表单', 'Apply to Form') }}</button>
              </div>
            </div>
          </section>
        </div>

        <aside class="space-y-4">
          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ localText('草稿状态', 'Draft Status') }}</h3>
            <dl class="mt-3 space-y-2 text-sm">
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">{{ localText('总开关', 'Enabled') }}</dt>
                <dd class="font-medium text-gray-900 dark:text-white">{{ draft.enabled ? localText('开启', 'On') : localText('关闭', 'Off') }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">{{ localText('教程数', 'Tutorials') }}</dt>
                <dd class="font-medium text-gray-900 dark:text-white">{{ draft.tutorials.length }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">FAQ</dt>
                <dd class="font-medium text-gray-900 dark:text-white">{{ draft.faqs.length }}</dd>
              </div>
            </dl>
          </div>

          <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ localText('已发布', 'Published') }}</h3>
            <dl class="mt-3 space-y-2 text-sm">
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">{{ localText('标题', 'Title') }}</dt>
                <dd class="truncate font-medium text-gray-900 dark:text-white">{{ published?.title || '-' }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">{{ localText('教程数', 'Tutorials') }}</dt>
                <dd class="font-medium text-gray-900 dark:text-white">{{ published?.tutorials?.length || 0 }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500 dark:text-gray-400">FAQ</dt>
                <dd class="font-medium text-gray-900 dark:text-white">{{ published?.faqs?.length || 0 }}</dd>
              </div>
            </dl>
          </div>

          <div class="rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/20 dark:text-amber-200">
            {{ localText('发布会覆盖用户侧帮助中心内容；用户没有编辑权限。附件上传后需挂到教程并保存/发布才会展示。', 'Publishing replaces the user-facing Help Center; users cannot edit it. Uploaded attachments must be attached to a tutorial and saved/published before they appear.') }}
          </div>
        </aside>
      </div>
    </section>

    <div v-if="imageDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4">
      <div class="w-full max-w-3xl rounded-lg bg-white shadow-xl dark:bg-dark-800">
        <div class="flex items-start justify-between gap-4 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ localText('插入步骤图片', 'Insert Step Image') }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('可以直接粘贴截图、拖拽图片，或从本地选择图片。', 'Paste a screenshot, drop an image, or choose a local file.') }}</p>
          </div>
          <button type="button" class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700" @click="closeImageDialog">
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="space-y-4 p-5">
          <label class="block">
            <span class="input-label">{{ localText('图片名称', 'Image Label') }}</span>
            <input v-model="imageDialog.label" class="input mt-1" :placeholder="localText('例如：Codex 设置页截图', 'Example: Codex settings screenshot')" />
          </label>
          <div
            class="flex min-h-[260px] flex-col items-center justify-center rounded-lg border border-dashed border-primary-300 bg-primary-50/40 p-5 text-center outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-200 dark:border-primary-800 dark:bg-primary-900/10"
            tabindex="0"
            contenteditable="true"
            @paste.prevent="handleImageDialogPaste"
            @dragover.prevent
            @drop.prevent="handleImageDialogDrop"
          >
            <img v-if="imageDialog.previewURL" :src="imageDialog.previewURL" :alt="imageDialog.label" class="max-h-[220px] max-w-full rounded-md border border-gray-200 bg-white object-contain dark:border-dark-700 dark:bg-dark-900" />
            <div v-else class="space-y-3">
              <Icon name="upload" size="lg" class="mx-auto text-primary-500" />
              <p class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ localText('点击此区域后按 Ctrl+V 粘贴图片', 'Click here and press Ctrl+V to paste an image') }}</p>
              <p class="text-xs text-gray-500 dark:text-gray-400">{{ localText('也可以把图片拖进来，或使用下面的本地选择。', 'You can also drop an image here or use the file picker below.') }}</p>
            </div>
          </div>
          <div class="flex flex-wrap items-center justify-between gap-3">
            <label class="btn btn-secondary cursor-pointer">
              <Icon name="upload" size="sm" />
              {{ localText('选择图片', 'Choose Image') }}
              <input type="file" accept="image/*" class="sr-only" @change="handleImageDialogFileInput" />
            </label>
            <p v-if="imageDialog.file" class="min-w-0 flex-1 truncate text-sm text-gray-500 dark:text-gray-400">{{ imageDialog.file.name }}</p>
          </div>
        </div>
        <div class="flex justify-end gap-2 border-t border-gray-100 px-5 py-4 dark:border-dark-700">
          <button type="button" class="btn btn-secondary" @click="closeImageDialog">{{ localText('取消', 'Cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="!imageDialog.file || uploadingAttachment" @click="confirmImageDialog">
            <Icon name="plus" size="sm" />
            {{ uploadingAttachment ? localText('上传中...', 'Uploading...') : localText('插入图片', 'Insert Image') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="codeDialog.open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/45 p-4">
      <div class="w-full max-w-4xl rounded-lg bg-white shadow-xl dark:bg-dark-800">
        <div class="flex items-start justify-between gap-4 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
          <div>
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ localText('插入代码块', 'Insert Code Block') }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ localText('可以直接粘贴命令或 Markdown 代码围栏。', 'Paste commands or a fenced Markdown code block.') }}</p>
          </div>
          <button type="button" class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700" @click="closeCodeBlockDialog">
            <Icon name="x" size="sm" />
          </button>
        </div>
        <div class="grid grid-cols-1 gap-4 p-5 lg:grid-cols-3">
          <label class="block">
            <span class="input-label">{{ localText('标题', 'Title') }}</span>
            <input v-model="codeDialog.title" class="input mt-1" :placeholder="localText('例如：环境变量配置', 'Example: Environment variables')" />
          </label>
          <label class="block">
            <span class="input-label">{{ localText('语言', 'Language') }}</span>
            <input v-model="codeDialog.language" class="input mt-1" placeholder="bash" />
          </label>
          <div class="flex items-end text-xs text-gray-500 dark:text-gray-400">
            {{ localText('如果粘贴 ```json 代码围栏，会自动识别语言并去掉围栏。', 'Fenced blocks such as ```json are parsed automatically.') }}
          </div>
          <label class="block lg:col-span-3">
            <span class="input-label">{{ localText('代码内容', 'Code Content') }}</span>
            <textarea v-model="codeDialog.content" spellcheck="false" class="input mt-1 min-h-[360px] font-mono text-sm leading-6" placeholder="```bash&#10;export API_KEY=...&#10;```"></textarea>
          </label>
        </div>
        <div class="flex justify-end gap-2 border-t border-gray-100 px-5 py-4 dark:border-dark-700">
          <button type="button" class="btn btn-secondary" @click="closeCodeBlockDialog">{{ localText('取消', 'Cancel') }}</button>
          <button type="button" class="btn btn-primary" :disabled="!codeDialog.content.trim()" @click="confirmCodeBlockDialog">
            <Icon name="plus" size="sm" />
            {{ localText('插入代码块', 'Insert Code Block') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { cloneHelpCenterTutorial, dataURLToFile, extractImageSourcesFromHTML, parseCodeBlockDraft } from '@/utils/helpCenterEditor'
import type { HelpCenterAttachment, HelpCenterCodeBlock, HelpCenterConfig, HelpCenterFAQ, HelpCenterStep, HelpCenterTutorial } from '@/types'

const EditableList = defineComponent({
  props: {
    title: { type: String, required: true },
    items: { type: Array, required: true },
    addLabel: { type: String, required: true },
  },
  emits: ['add', 'remove'],
  setup(props, { emit, slots }) {
    return () => h('div', { class: 'rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800' }, [
      h('div', { class: 'flex items-center justify-between gap-3' }, [
        h('h4', { class: 'text-sm font-semibold text-gray-900 dark:text-white' }, props.title),
        h('button', { type: 'button', class: 'btn btn-secondary text-xs', onClick: () => emit('add') }, props.addLabel),
      ]),
      h('div', { class: 'mt-3 space-y-3' }, [
        ...(props.items as unknown[]).map((item, index) => h('div', { class: 'rounded-md border border-gray-100 p-3 dark:border-dark-700' }, [
          h('div', { class: 'flex justify-end pb-2' }, [
            h('button', { type: 'button', class: 'btn btn-danger text-xs', onClick: () => emit('remove', index) }, '删除'),
          ]),
          slots.default?.({ item, index }),
        ])),
      ]),
    ])
  },
})

const { locale } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const publishing = ref(false)
const uploadingAttachment = ref(false)
const draft = ref<HelpCenterConfig | null>(null)
const published = ref<HelpCenterConfig | null>(null)
const selectedTutorialIndex = ref(0)
const showAdvancedJson = ref(false)
const advancedJson = ref('')
const jsonError = ref('')
const imageDialog = ref<{
  open: boolean
  step: HelpCenterStep | null
  label: string
  file: File | null
  previewURL: string
}>({
  open: false,
  step: null,
  label: '',
  file: null,
  previewURL: '',
})
const codeDialog = ref<{
  open: boolean
  target: HelpCenterCodeBlock[] | null
  contextTitle: string
  title: string
  language: string
  content: string
}>({
  open: false,
  target: null,
  contextTitle: '',
  title: '',
  language: 'bash',
  content: '',
})

const busy = computed(() => loading.value || saving.value || publishing.value)
const selectedTutorial = computed(() => draft.value?.tutorials[selectedTutorialIndex.value] || null)

watch(draft, (value) => {
  if (value) advancedJson.value = formatConfig(value)
}, { deep: true })

function localText(zh: string, en: string): string {
  return locale.value.startsWith('zh') ? zh : en
}

function formatConfig(config: HelpCenterConfig): string {
  return JSON.stringify(config, null, 2)
}

function normalizeDraft(config: HelpCenterConfig): HelpCenterConfig {
  return {
    enabled: !!config.enabled,
    base_url: config.base_url || '',
    title: config.title || '',
    description: config.description || '',
    key_created_prompt: {
      enabled: !!config.key_created_prompt?.enabled,
      title: config.key_created_prompt?.title || '',
      description: config.key_created_prompt?.description || '',
      primary_action_label: config.key_created_prompt?.primary_action_label || '',
      primary_action_url: config.key_created_prompt?.primary_action_url || '/help-center',
      secondary_action_label: config.key_created_prompt?.secondary_action_label || '',
      secondary_action_url: config.key_created_prompt?.secondary_action_url || '/keys',
      dismiss_label: config.key_created_prompt?.dismiss_label || '',
    },
    tutorials: (config.tutorials || []).map(normalizeTutorial),
    faqs: (config.faqs || []).map(normalizeFAQ),
  }
}

function normalizeTutorial(tutorial: Partial<HelpCenterTutorial>): HelpCenterTutorial {
  return {
    id: tutorial.id || uniqueID('tutorial'),
    enabled: tutorial.enabled ?? true,
    sort_order: tutorial.sort_order ?? nextSort(draft.value?.tutorials || []),
    title: tutorial.title || '',
    badge: tutorial.badge || '',
    summary: tutorial.summary || '',
    content_md: tutorial.content_md || '',
    steps: (tutorial.steps || []).map(normalizeStep),
    code_blocks: normalizeCodeBlocks(tutorial.code_blocks || []),
    links: (tutorial.links || []).filter((link) => link.label || link.url),
    attachments: normalizeAttachments(tutorial.attachments || []),
  }
}

function normalizeStep(step: Partial<HelpCenterStep>): HelpCenterStep {
  return {
    title: step.title || '',
    description: step.description || '',
    code_blocks: normalizeCodeBlocks(step.code_blocks || []),
    images: normalizeAttachments(step.images || []),
    attachments: normalizeAttachments(step.attachments || []),
  }
}

function normalizeCodeBlocks(blocks: HelpCenterCodeBlock[]): HelpCenterCodeBlock[] {
  return blocks
    .filter((block) => block.title || block.content || (block.language && block.language !== 'bash'))
    .map((block) => ({
      title: block.title || '',
      language: block.language || 'bash',
      content: block.content || '',
    }))
}

function normalizeAttachments(attachments: HelpCenterAttachment[]): HelpCenterAttachment[] {
  return attachments
    .filter((attachment) => attachment.label || attachment.url || attachment.file_name)
    .map((attachment) => ({
      label: attachment.label || attachment.file_name || attachment.url || '',
      url: attachment.url || '',
      file_name: attachment.file_name || '',
    }))
}

function emptyStep(): HelpCenterStep {
  return {
    title: '',
    description: '',
    code_blocks: [],
    images: [],
    attachments: [],
  }
}

function normalizeFAQ(faq: Partial<HelpCenterFAQ>): HelpCenterFAQ {
  return {
    id: faq.id || uniqueID('faq'),
    enabled: faq.enabled ?? true,
    sort_order: faq.sort_order ?? nextSort(draft.value?.faqs || []),
    question: faq.question || '',
    answer_md: faq.answer_md || '',
    tags: faq.tags || [],
  }
}

function nextSort(items: Array<{ sort_order: number }>): number {
  return items.length ? Math.max(...items.map((item) => Number(item.sort_order) || 0)) + 10 : 10
}

function uniqueID(prefix: string): string {
  return `${prefix}-${Date.now().toString(36)}`
}

function parseTags(value: string): string[] {
  return value.split(',').map((item) => item.trim()).filter(Boolean)
}

function updateFAQTags(faq: HelpCenterFAQ, event: Event): void {
  const input = event.target as HTMLInputElement | null
  faq.tags = parseTags(input?.value || '')
}

function currentConfigForSave(): HelpCenterConfig | null {
  if (!draft.value) return null
  return normalizeDraft(JSON.parse(JSON.stringify(draft.value)) as HelpCenterConfig)
}

async function loadConfig(): Promise<void> {
  loading.value = true
  jsonError.value = ''
  try {
    const response = await adminAPI.helpCenter.get()
    draft.value = normalizeDraft(response.draft)
    published.value = normalizeDraft(response.published)
    selectedTutorialIndex.value = 0
  } catch (error) {
    appStore.showError(localText('帮助中心配置加载失败', 'Failed to load Help Center config'))
  } finally {
    loading.value = false
  }
}

async function saveDraftOnly(): Promise<HelpCenterConfig | null> {
  const config = currentConfigForSave()
  if (!config) return null
  saving.value = true
  try {
    const saved = await adminAPI.helpCenter.saveDraft(config)
    draft.value = normalizeDraft(saved)
    appStore.showSuccess(localText('草稿已保存', 'Draft saved'))
    return draft.value
  } catch (error: any) {
    appStore.showError(error?.message || localText('草稿保存失败', 'Failed to save draft'))
    return null
  } finally {
    saving.value = false
  }
}

async function publishCurrentDraft(): Promise<void> {
  const saved = await saveDraftOnly()
  if (!saved) return
  publishing.value = true
  try {
    published.value = normalizeDraft(await adminAPI.helpCenter.publishDraft())
    appStore.showSuccess(localText('帮助中心已发布', 'Help Center published'))
  } catch (error: any) {
    appStore.showError(error?.message || localText('发布失败', 'Failed to publish'))
  } finally {
    publishing.value = false
  }
}

function addTutorial(): void {
  if (!draft.value) return
  draft.value.tutorials.push(normalizeTutorial({
    id: uniqueID('client'),
    sort_order: nextSort(draft.value.tutorials),
    title: localText('新教程', 'New Tutorial'),
    badge: 'Guide',
  }))
  selectedTutorialIndex.value = draft.value.tutorials.length - 1
}

function moveTutorial(delta: number): void {
  if (!draft.value) return
  const from = selectedTutorialIndex.value
  const to = from + delta
  if (to < 0 || to >= draft.value.tutorials.length) return
  const [item] = draft.value.tutorials.splice(from, 1)
  draft.value.tutorials.splice(to, 0, item)
  selectedTutorialIndex.value = to
}

function duplicateTutorial(): void {
  if (!draft.value || !selectedTutorial.value) return
  const copy = normalizeTutorial(cloneHelpCenterTutorial(
    selectedTutorial.value,
    draft.value.tutorials,
    nextSort(draft.value.tutorials),
    localText('副本', 'Copy'),
  ))
  draft.value.tutorials.splice(selectedTutorialIndex.value + 1, 0, copy)
  selectedTutorialIndex.value += 1
}

function removeTutorial(): void {
  if (!draft.value) return
  draft.value.tutorials.splice(selectedTutorialIndex.value, 1)
  selectedTutorialIndex.value = Math.max(0, selectedTutorialIndex.value - 1)
}

function addFAQ(): void {
  if (!draft.value) return
  draft.value.faqs.push(normalizeFAQ({
    id: uniqueID('faq'),
    sort_order: nextSort(draft.value.faqs),
    question: localText('新问题', 'New Question'),
    answer_md: '',
    tags: [],
  }))
}

function moveFAQ(index: number, delta: number): void {
  if (!draft.value) return
  const to = index + delta
  if (to < 0 || to >= draft.value.faqs.length) return
  const [item] = draft.value.faqs.splice(index, 1)
  draft.value.faqs.splice(to, 0, item)
}

async function handleAttachmentUpload(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !selectedTutorial.value) return
  uploadingAttachment.value = true
  try {
    const attachment = await adminAPI.helpCenter.uploadAttachment(file)
    selectedTutorial.value.attachments.push(attachment)
    appStore.showSuccess(localText('附件已上传', 'Attachment uploaded'))
  } catch (error: any) {
    appStore.showError(error?.message || localText('附件上传失败', 'Failed to upload attachment'))
  } finally {
    uploadingAttachment.value = false
  }
}

function revokeImageDialogPreview(): void {
  if (imageDialog.value.previewURL) {
    URL.revokeObjectURL(imageDialog.value.previewURL)
  }
  imageDialog.value.previewURL = ''
}

function openImageDialog(step: HelpCenterStep): void {
  revokeImageDialogPreview()
  imageDialog.value = {
    open: true,
    step,
    label: '',
    file: null,
    previewURL: '',
  }
}

function closeImageDialog(): void {
  revokeImageDialogPreview()
  imageDialog.value.open = false
  imageDialog.value.step = null
  imageDialog.value.file = null
}

function setImageDialogFile(file: File): void {
  const isImage = file.type.startsWith('image/') || /\.(png|jpe?g|gif|webp|avif|svg)$/i.test(file.name)
  if (!isImage) {
    appStore.showError(localText('请粘贴或选择图片文件', 'Please paste or choose an image file'))
    return
  }
  revokeImageDialogPreview()
  imageDialog.value.file = file
  imageDialog.value.previewURL = URL.createObjectURL(file)
  if (!imageDialog.value.label) {
    imageDialog.value.label = file.name.replace(/\.[^.]+$/, '')
  }
}

async function handleImageDialogPaste(event: ClipboardEvent): Promise<void> {
  const imageFile = Array.from(event.clipboardData?.files || []).find((file) => file.type.startsWith('image/'))
  if (imageFile) {
    setImageDialogFile(imageFile)
    return
  }

  const html = event.clipboardData?.getData('text/html') || ''
  const dataURL = extractImageSourcesFromHTML(html).find((src) => src.startsWith('data:image/'))
  if (dataURL) {
    setImageDialogFile(await dataURLToFile(dataURL, `help-center-image-${Date.now()}`))
    return
  }

  appStore.showError(localText('剪贴板里没有可用图片', 'No image found in the clipboard'))
}

function handleImageDialogDrop(event: DragEvent): void {
  const imageFile = Array.from(event.dataTransfer?.files || []).find((file) => file.type.startsWith('image/'))
  if (!imageFile) {
    appStore.showError(localText('请拖入图片文件', 'Please drop an image file'))
    return
  }
  setImageDialogFile(imageFile)
}

function handleImageDialogFileInput(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (file) setImageDialogFile(file)
}

async function confirmImageDialog(): Promise<void> {
  const step = imageDialog.value.step
  const file = imageDialog.value.file
  if (!step || !file) return
  uploadingAttachment.value = true
  try {
    const attachment = await adminAPI.helpCenter.uploadAttachment(file)
    if (!step.images) step.images = []
    step.images.push({
      ...attachment,
      label: imageDialog.value.label.trim() || attachment.label,
    })
    appStore.showSuccess(localText('图片已插入', 'Image inserted'))
    closeImageDialog()
  } catch (error: any) {
    appStore.showError(error?.message || localText('图片上传失败', 'Failed to upload image'))
  } finally {
    uploadingAttachment.value = false
  }
}

function openCodeBlockDialog(target: HelpCenterCodeBlock[], contextTitle: string): void {
  codeDialog.value = {
    open: true,
    target,
    contextTitle,
    title: contextTitle ? `${contextTitle} ${localText('代码', 'Code')}` : localText('代码块', 'Code Block'),
    language: 'bash',
    content: '',
  }
}

function closeCodeBlockDialog(): void {
  codeDialog.value.open = false
  codeDialog.value.target = null
}

function confirmCodeBlockDialog(): void {
  if (!codeDialog.value.target) return
  const parsed = parseCodeBlockDraft(codeDialog.value.content, codeDialog.value.language || 'bash')
  codeDialog.value.target.push({
    title: codeDialog.value.title.trim() || localText('代码块', 'Code Block'),
    language: parsed.language,
    content: parsed.content,
  })
  closeCodeBlockDialog()
}

async function handleStepImageUpload(step: HelpCenterStep, event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  const isImage = file.type.startsWith('image/') || /\.(png|jpe?g|gif|webp|avif|svg)$/i.test(file.name)
  if (!isImage) {
    appStore.showError(localText('请上传图片文件', 'Please upload an image file'))
    return
  }
  uploadingAttachment.value = true
  try {
    const attachment = await adminAPI.helpCenter.uploadAttachment(file)
    if (!step.images) step.images = []
    step.images.push(attachment)
    appStore.showSuccess(localText('图片已上传', 'Image uploaded'))
  } catch (error: any) {
    appStore.showError(error?.message || localText('图片上传失败', 'Failed to upload image'))
  } finally {
    uploadingAttachment.value = false
  }
}

async function handleStepAttachmentUpload(step: HelpCenterStep, event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  uploadingAttachment.value = true
  try {
    const attachment = await adminAPI.helpCenter.uploadAttachment(file)
    if (!step.attachments) step.attachments = []
    step.attachments.push(attachment)
    appStore.showSuccess(localText('附件已上传', 'Attachment uploaded'))
  } catch (error: any) {
    appStore.showError(error?.message || localText('附件上传失败', 'Failed to upload attachment'))
  } finally {
    uploadingAttachment.value = false
  }
}

function applyAdvancedJson(): void {
  jsonError.value = ''
  try {
    const parsed = JSON.parse(advancedJson.value) as HelpCenterConfig
    draft.value = normalizeDraft(parsed)
    selectedTutorialIndex.value = 0
  } catch (error) {
    jsonError.value = error instanceof Error ? error.message : localText('JSON 格式错误', 'Invalid JSON')
  }
}

onMounted(() => {
  loadConfig()
})
</script>
