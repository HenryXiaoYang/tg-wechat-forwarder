import { forwardRef, type ButtonHTMLAttributes } from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "../../lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-[10px] text-sm font-medium transition-colors outline-none focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--surface)] disabled:pointer-events-none disabled:opacity-45",
  {
    variants: {
      variant: {
        default: "bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)]",
        outline: "border border-border bg-surface hover:bg-muted",
        ghost: "text-muted-foreground hover:bg-muted hover:text-foreground",
        danger: "text-red-600 hover:bg-red-500/10 dark:text-red-400",
      },
      size: { default: "h-10 px-5", sm: "h-8 px-3.5 text-xs", icon: "size-9 rounded-full" },
    },
    defaultVariants: { variant: "default", size: "default" },
  },
)

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {}
export const Button = forwardRef<HTMLButtonElement, ButtonProps>(({ className, variant, size, ...props }, ref) => (
  <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} {...props} />
))
Button.displayName = "Button"
