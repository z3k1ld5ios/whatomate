<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { campaignsService, api } from '@/services/api'
import { toast } from 'vue-sonner'
import { useUnsavedChangesGuard } from '@/composables/useUnsavedChangesGuard'
import DetailPageLayout from '@/components/shared/DetailPageLayout.vue'
import MetadataPanel from '@/components/shared/MetadataPanel.vue'
import AuditLogPanel from '@/components/shared/AuditLogPanel.vue'
import UnsavedChangesDialog from '@/components/shared/UnsavedChangesDialog.vue'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Megaphone,
  Trash2,
  Save,
  CheckCircle,
  Play,
  Pause,
  Clock,
  AlertCircle,
  Users,
  Send,
  Eye,
  XCircle,
} from 'lucide-vue-next'

interface Campaign {
  id: string
  name: string
  whatsapp_account?: string
  template_id?: string
  template_name?: string
  header_media_id?: string
  header_media_filename?: string
  header_media_mime_type?: string
  status: string
  total_recipients: number
  sent_count: number
  delivered_count: number
  read_count: number
  failed_count: number
  scheduled_at?: string
  started_at?: string
  completed_at?: string
  created_by_name?: string
  updated_by_name?: string
  created_at: string
  updated_at: string
}

interface Account {
  id: string
  name: string
}

interface Template {
  id: string
  name: string
  display_name?: string
  status: string
}

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const campaignId = computed(() => route.params.id as string)
const isNew = computed(() => campaignId.value === 'new')
const campaign = ref<Campaign | null>(null)
const isLoading = ref(true)
const isNotFound = ref(false)
const isSaving = ref(false)
const hasChanges = ref(false)
const deleteDialogOpen = ref(false)

const accounts = ref<Account[]>([])
const templates = ref<Template[]>([])

const { showLeaveDialog, confirmLeave, cancelLeave } = useUnsavedChangesGuard(hasChanges)

const isDraft = computed(() => isNew.value || campaign.value?.status === 'draft')

const form = ref({
  name: '',
  whatsapp_account: '',
  template_id: '',
  scheduled_at: '',
})

const breadcrumbs = computed(() => [
  { label: t('nav.campaigns', 'Campaigns'), href: '/campaigns' },
  { label: isNew.value ? t('campaigns.newCampaign', 'New Campaign') : (campaign.value?.name || '') },
])

// Status helpers
function getStatusIcon(status: string) {
  switch (status) {
    case 'completed':
      return CheckCircle
    case 'running':
    case 'processing':
    case 'queued':
      return Play
    case 'paused':
      return Pause
    case 'scheduled':
      return Clock
    case 'failed':
    case 'cancelled':
      return AlertCircle
    default:
      return Megaphone
  }
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'border-green-600 text-green-600'
    case 'running':
    case 'processing':
    case 'queued':
      return 'border-blue-600 text-blue-600'
    case 'failed':
    case 'cancelled':
      return 'border-destructive text-destructive'
    default:
      return ''
  }
}

async function loadAccounts() {
  try {
    const response = await api.get('/accounts')
    accounts.value = (response.data as any).data || response.data || []
  } catch {
    // Silently fail — accounts list is non-critical
  }
}

async function loadTemplates() {
  if (!form.value.whatsapp_account) {
    templates.value = []
    return
  }
  try {
    const response = await api.get('/templates', {
      params: { whatsapp_account: form.value.whatsapp_account },
    })
    templates.value = (response.data as any).data || response.data || []
  } catch {
    templates.value = []
  }
}

async function loadCampaign() {
  isLoading.value = true
  isNotFound.value = false
  try {
    const response = await campaignsService.get(campaignId.value)
    const data = (response.data as any).data || response.data
    campaign.value = data
    syncForm()
    nextTick(() => { hasChanges.value = false })
  } catch {
    isNotFound.value = true
  } finally {
    isLoading.value = false
  }
}

function syncForm() {
  if (!campaign.value) return
  form.value = {
    name: campaign.value.name || '',
    whatsapp_account: campaign.value.whatsapp_account || '',
    template_id: campaign.value.template_id || '',
    scheduled_at: campaign.value.scheduled_at ? campaign.value.scheduled_at.slice(0, 16) : '',
  }
}

// Track form changes
watch(form, () => {
  hasChanges.value = true
}, { deep: true })

// Reload templates when account changes
watch(() => form.value.whatsapp_account, (newVal, oldVal) => {
  if (newVal !== oldVal) {
    // Clear template selection if account changed
    if (oldVal) {
      form.value.template_id = ''
    }
    loadTemplates()
  }
})

async function save() {
  if (!form.value.name.trim()) {
    toast.error(t('campaigns.nameRequired', 'Campaign name is required'))
    return
  }
  isSaving.value = true
  try {
    const payload: Record<string, any> = {
      name: form.value.name,
      whatsapp_account: form.value.whatsapp_account || undefined,
      template_id: form.value.template_id || undefined,
      scheduled_at: form.value.scheduled_at || undefined,
    }
    if (isNew.value) {
      const response = await campaignsService.create(payload)
      const created = (response.data as any).data || response.data
      hasChanges.value = false
      toast.success(t('campaigns.created', 'Campaign created'))
      router.replace(`/campaigns/${created.id}`)
    } else {
      await campaignsService.update(campaign.value!.id, payload)
      await loadCampaign()
      hasChanges.value = false
      toast.success(t('campaigns.updated', 'Campaign updated'))
    }
  } catch {
    toast.error(
      isNew.value
        ? t('campaigns.createFailed', 'Failed to create campaign')
        : t('campaigns.updateFailed', 'Failed to update campaign'),
    )
  } finally {
    isSaving.value = false
  }
}

async function deleteCampaign() {
  if (!campaign.value) return
  try {
    await campaignsService.delete(campaign.value.id)
    toast.success(t('campaigns.deleted', 'Campaign deleted'))
    router.push('/campaigns')
  } catch {
    toast.error(t('campaigns.deleteFailed', 'Failed to delete campaign'))
  }
  deleteDialogOpen.value = false
}

onMounted(async () => {
  await loadAccounts()
  if (isNew.value) {
    isLoading.value = false
    hasChanges.value = false
  } else {
    await loadCampaign()
    // Load templates for the selected account after campaign loads
    if (form.value.whatsapp_account) {
      await loadTemplates()
    }
  }
})
</script>

<template>
  <div class="h-full">
  <DetailPageLayout
    :title="isNew ? $t('campaigns.newCampaign', 'New Campaign') : (campaign?.name || '')"
    :icon="Megaphone"
    icon-gradient="bg-gradient-to-br from-pink-500 to-rose-600 shadow-pink-500/20"
    back-link="/campaigns"
    :breadcrumbs="breadcrumbs"
    :is-loading="isLoading"
    :is-not-found="isNotFound"
    :not-found-title="$t('campaigns.notFound', 'Campaign not found')"
  >
    <template #actions>
      <div class="flex items-center gap-2">
        <!-- Status badge for existing campaigns -->
        <Badge
          v-if="!isNew && campaign"
          variant="outline"
          :class="[getStatusClass(campaign.status), 'text-xs']"
        >
          <component :is="getStatusIcon(campaign.status)" class="h-3 w-3 mr-1" />
          {{ campaign.status }}
        </Badge>

        <Button
          v-if="isDraft && (hasChanges || isNew)"
          size="sm"
          @click="save"
          :disabled="isSaving"
        >
          <Save class="h-4 w-4 mr-1" />
          {{ isSaving ? $t('common.saving', 'Saving...') : isNew ? $t('common.create') : $t('common.save') }}
        </Button>
        <Button
          v-if="isDraft && !isNew"
          variant="destructive"
          size="sm"
          @click="deleteDialogOpen = true"
        >
          <Trash2 class="h-4 w-4 mr-1" /> {{ $t('common.delete') }}
        </Button>
      </div>
    </template>

    <!-- Details Card -->
    <Card>
      <CardHeader class="pb-3">
        <CardTitle class="text-sm font-medium">{{ $t('campaigns.details', 'Details') }}</CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-1.5">
          <Label class="text-xs">{{ $t('campaigns.name', 'Name') }} *</Label>
          <Input v-model="form.name" :disabled="!isDraft" :placeholder="$t('campaigns.namePlaceholder', 'Enter campaign name')" />
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs">{{ $t('campaigns.whatsappAccount', 'WhatsApp Account') }}</Label>
          <Select v-model="form.whatsapp_account" :disabled="!isDraft">
            <SelectTrigger>
              <SelectValue :placeholder="$t('campaigns.selectAccount', 'Select account')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="account in accounts" :key="account.id" :value="account.id">
                {{ account.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs">{{ $t('campaigns.template', 'Template') }}</Label>
          <Select v-model="form.template_id" :disabled="!isDraft || !form.whatsapp_account">
            <SelectTrigger>
              <SelectValue :placeholder="form.whatsapp_account ? $t('campaigns.selectTemplate', 'Select template') : $t('campaigns.selectAccountFirst', 'Select an account first')" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="tmpl in templates" :key="tmpl.id" :value="tmpl.id">
                {{ tmpl.display_name || tmpl.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs">{{ $t('campaigns.scheduledAt', 'Schedule') }}</Label>
          <Input v-model="form.scheduled_at" type="datetime-local" :disabled="!isDraft" />
        </div>
        <div v-if="!isNew && campaign" class="space-y-1.5">
          <Label class="text-xs">{{ $t('campaigns.status', 'Status') }}</Label>
          <div>
            <Badge variant="outline" :class="[getStatusClass(campaign.status), 'text-xs']">
              <component :is="getStatusIcon(campaign.status)" class="h-3 w-3 mr-1" />
              {{ campaign.status }}
            </Badge>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Stats Card (existing campaigns only) -->
    <Card v-if="!isNew && campaign">
      <CardHeader class="pb-3">
        <CardTitle class="text-sm font-medium">{{ $t('campaigns.statistics', 'Statistics') }}</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
          <div class="flex flex-col items-center gap-1 rounded-lg border p-3">
            <Users class="h-4 w-4 text-muted-foreground" />
            <span class="text-lg font-semibold">{{ campaign.total_recipients }}</span>
            <span class="text-[10px] text-muted-foreground uppercase tracking-wide">{{ $t('campaigns.totalRecipients', 'Recipients') }}</span>
          </div>
          <div class="flex flex-col items-center gap-1 rounded-lg border p-3">
            <Send class="h-4 w-4 text-blue-500" />
            <span class="text-lg font-semibold">{{ campaign.sent_count }}</span>
            <span class="text-[10px] text-muted-foreground uppercase tracking-wide">{{ $t('campaigns.sent', 'Sent') }}</span>
          </div>
          <div class="flex flex-col items-center gap-1 rounded-lg border p-3">
            <CheckCircle class="h-4 w-4 text-green-500" />
            <span class="text-lg font-semibold">{{ campaign.delivered_count }}</span>
            <span class="text-[10px] text-muted-foreground uppercase tracking-wide">{{ $t('campaigns.delivered', 'Delivered') }}</span>
          </div>
          <div class="flex flex-col items-center gap-1 rounded-lg border p-3">
            <Eye class="h-4 w-4 text-purple-500" />
            <span class="text-lg font-semibold">{{ campaign.read_count }}</span>
            <span class="text-[10px] text-muted-foreground uppercase tracking-wide">{{ $t('campaigns.read', 'Read') }}</span>
          </div>
          <div class="flex flex-col items-center gap-1 rounded-lg border p-3">
            <XCircle class="h-4 w-4 text-destructive" />
            <span class="text-lg font-semibold">{{ campaign.failed_count }}</span>
            <span class="text-[10px] text-muted-foreground uppercase tracking-wide">{{ $t('campaigns.failed', 'Failed') }}</span>
          </div>
        </div>

        <!-- Progress Bar -->
        <div v-if="campaign.total_recipients > 0" class="mt-4 space-y-2">
          <div class="flex items-center justify-between text-xs text-muted-foreground">
            <span>{{ $t('campaigns.progress', 'Progress') }}</span>
            <span>{{ Math.round(((campaign.sent_count + campaign.failed_count) / campaign.total_recipients) * 100) }}%</span>
          </div>
          <div class="h-2.5 w-full bg-muted rounded-full overflow-hidden flex">
            <div
              class="bg-green-500 h-full transition-all duration-500"
              :style="{ width: `${(campaign.delivered_count / campaign.total_recipients) * 100}%` }"
            />
            <div
              class="bg-blue-500 h-full transition-all duration-500"
              :style="{ width: `${((campaign.sent_count - campaign.delivered_count) / campaign.total_recipients) * 100}%` }"
            />
            <div
              class="bg-destructive h-full transition-all duration-500"
              :style="{ width: `${(campaign.failed_count / campaign.total_recipients) * 100}%` }"
            />
          </div>
          <div class="flex items-center gap-4 text-[10px] text-muted-foreground">
            <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-green-500" /> {{ $t('campaigns.delivered', 'Delivered') }}</span>
            <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-blue-500" /> {{ $t('campaigns.sent', 'Sent') }}</span>
            <span class="flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-destructive" /> {{ $t('campaigns.failed', 'Failed') }}</span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Audit Log -->
    <AuditLogPanel
      v-if="campaign && !isNew"
      resource-type="campaign"
      :resource-id="campaign.id"
    />

    <!-- Sidebar -->
    <template v-if="!isNew" #sidebar>
      <MetadataPanel
        :created-at="campaign?.created_at"
        :updated-at="campaign?.updated_at"
        :created-by-name="campaign?.created_by_name"
        :updated-by-name="campaign?.updated_by_name"
      />
    </template>
  </DetailPageLayout>

  <!-- Delete Confirmation -->
  <AlertDialog v-model:open="deleteDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ $t('campaigns.deleteCampaign', 'Delete Campaign') }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ $t('campaigns.deleteConfirm', 'Are you sure? This action cannot be undone.') }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ $t('common.cancel') }}</AlertDialogCancel>
        <AlertDialogAction @click="deleteCampaign">{{ $t('common.delete') }}</AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>

  <UnsavedChangesDialog :open="showLeaveDialog" @stay="cancelLeave" @leave="confirmLeave" />
  </div>
</template>
