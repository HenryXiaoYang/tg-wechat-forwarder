import { forwardRef, type InputHTMLAttributes } from "react"
import { cn } from "../../lib/utils"

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(({ className, ...props }, ref) => (
  <input ref={ref} className={cn("flex h-10 w-full rounded-[10px] border border-transparent bg-muted px-3.5 text-sm text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-[var(--accent)] focus:bg-surface disabled:opacity-50", className)} {...props} />
))
Input.displayName = "Input"
