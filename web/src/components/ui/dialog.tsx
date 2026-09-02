import type { ComponentProps, HTMLAttributes } from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { X } from "lucide-react"
import { cn } from "../../lib/utils"

export const Dialog = DialogPrimitive.Root
export function DialogContent({ className, children, ...props }: ComponentProps<typeof DialogPrimitive.Content>) {
  return <DialogPrimitive.Portal>
    <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/40" />
    <DialogPrimitive.Content className={cn("fixed left-1/2 top-1/2 z-50 max-h-[88vh] w-[calc(100%-24px)] max-w-2xl -translate-x-1/2 -translate-y-1/2 overflow-y-auto rounded-[14px] bg-background p-0 text-foreground shadow-[var(--panel-shadow)] outline-none", className)} {...props}>
      {children}
      <DialogPrimitive.Close className="absolute right-4 top-4 grid size-8 place-items-center rounded-full text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-[var(--accent)]" aria-label="关闭"><X className="size-4" /></DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </DialogPrimitive.Portal>
}
export const DialogHeader = ({ className, ...props }: HTMLAttributes<HTMLDivElement>) => <div className={cn("border-b border-border px-6 py-5", className)} {...props} />
export const DialogTitle = DialogPrimitive.Title
export const DialogDescription = DialogPrimitive.Description
