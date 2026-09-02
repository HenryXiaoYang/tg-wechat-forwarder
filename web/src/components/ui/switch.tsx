import type { ComponentProps } from "react"
import * as SwitchPrimitive from "@radix-ui/react-switch"
import { cn } from "../../lib/utils"

export function Switch({ className, ...props }: ComponentProps<typeof SwitchPrimitive.Root>) {
  return <SwitchPrimitive.Root className={cn("inline-flex h-[22px] w-[38px] shrink-0 cursor-pointer items-center rounded-full bg-[#c4c9cc] outline-none transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:ring-offset-2 focus-visible:ring-offset-[var(--surface)] data-[state=checked]:bg-[var(--accent)] dark:bg-[#4b5b6b]", className)} {...props}>
    <SwitchPrimitive.Thumb className="pointer-events-none block size-[18px] translate-x-0.5 rounded-full bg-white shadow-[0_1px_3px_rgba(0,0,0,.3)] transition-transform data-[state=checked]:translate-x-[18px]" />
  </SwitchPrimitive.Root>
}
