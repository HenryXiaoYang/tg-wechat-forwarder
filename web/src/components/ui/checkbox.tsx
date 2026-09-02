import type { ComponentProps } from "react"
import * as CheckboxPrimitive from "@radix-ui/react-checkbox"
import { Check } from "lucide-react"
import { cn } from "../../lib/utils"

export function Checkbox({ className, ...props }: ComponentProps<typeof CheckboxPrimitive.Root>) {
  return <CheckboxPrimitive.Root className={cn("peer grid size-[22px] shrink-0 place-items-center rounded-full border-2 border-[#c4c9cc] bg-transparent outline-none transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--surface)] data-[state=checked]:border-[var(--accent)] data-[state=checked]:bg-[var(--accent)] data-[state=checked]:text-white dark:border-[#4b5b6b]", className)} {...props}>
    <CheckboxPrimitive.Indicator className="flex items-center justify-center"><Check className="size-3.5" strokeWidth={3.5} /></CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
}
