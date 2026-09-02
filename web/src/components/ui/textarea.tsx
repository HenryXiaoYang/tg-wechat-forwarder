import { forwardRef, type TextareaHTMLAttributes } from "react"
import { cn } from "../../lib/utils"

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(({ className, ...props }, ref) => (
  <textarea ref={ref} className={cn("flex min-h-24 w-full resize-y rounded-[10px] border border-transparent bg-muted px-3.5 py-2.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-[var(--accent)] focus:bg-surface", className)} {...props} />
))
Textarea.displayName = "Textarea"
