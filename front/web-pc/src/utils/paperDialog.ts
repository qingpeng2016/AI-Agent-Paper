import { ElMessageBox, type ElMessageBoxOptions } from 'element-plus'

export type PaperDialogVariant = 'default' | 'warning' | 'danger'

type PaperConfirmOptions = {
  confirmButtonText?: string
  cancelButtonText?: string
  variant?: PaperDialogVariant
}

type PaperPromptOptions = PaperConfirmOptions & {
  inputValue?: string
  inputPlaceholder?: string
}

function dialogClass(variant: PaperDialogVariant, extra?: string) {
  return ['paper-message-box', `paper-message-box--${variant}`, extra].filter(Boolean).join(' ')
}

const baseOptions = (): Partial<ElMessageBoxOptions> => ({
  customClass: dialogClass('default'),
  confirmButtonClass: 'paper-message-box__confirm',
  cancelButtonClass: 'paper-message-box__cancel',
  distinguishCancelAndClose: true,
  autofocus: false,
  closeOnClickModal: false,
  closeOnPressEscape: true,
})

export async function paperConfirm(
  message: string,
  title: string,
  options: PaperConfirmOptions = {},
): Promise<void> {
  const variant = options.variant ?? 'warning'
  await ElMessageBox.confirm(message, title, {
    ...baseOptions(),
    customClass: dialogClass(variant),
    confirmButtonText: options.confirmButtonText ?? '确定',
    cancelButtonText: options.cancelButtonText ?? '取消',
    type: variant === 'default' ? 'info' : 'warning',
    showClose: true,
  })
}

export async function paperPrompt(
  message: string,
  title: string,
  options: PaperPromptOptions = {},
): Promise<string> {
  const variant = options.variant ?? 'default'
  const { value } = await ElMessageBox.prompt(message, title, {
    ...baseOptions(),
    customClass: dialogClass(variant),
    confirmButtonText: options.confirmButtonText ?? '确定',
    cancelButtonText: options.cancelButtonText ?? '取消',
    inputValue: options.inputValue ?? '',
    inputPlaceholder: options.inputPlaceholder ?? '',
    showClose: true,
  })
  return value?.trim() ?? ''
}
