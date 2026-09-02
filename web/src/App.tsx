import { useCallback, useEffect, useMemo, useState, type Dispatch, type FormEvent, type ReactNode, type SetStateAction } from "react"
import {
  Archive, Check, ChevronRight, CircleAlert, Filter, Image, LoaderCircle, LockKeyhole,
  LogOut, MessageCircle, Moon, Pin, RefreshCw, Search, Send, Settings, ShieldCheck,
  Sun, Wifi, WifiOff,
} from "lucide-react"
import { Button } from "./components/ui/button"
import { Checkbox } from "./components/ui/checkbox"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "./components/ui/dialog"
import { Input } from "./components/ui/input"
import { Switch } from "./components/ui/switch"
import { Textarea } from "./components/ui/textarea"
import { cn } from "./lib/utils"

type TelegramState = { status: string; account?: string; qrCode?: string; qrExpires?: number; error?: string }
type DialogItem = {
  peerKey: string; kind: string; title: string; subtitle?: string; username?: string
  lastSender?: string; lastText?: string; selectable: boolean
  selected: boolean; adFilter: boolean; archived: boolean; pinned: boolean
}
type Delivery = { shortCode?: string; title: string; content?: string; time: string; status: string; error?: string }
type Dashboard = {
  version: string; telegram: TelegramState; dialogs: DialogItem[]
  pushplus: { topic: string; tokenConfigured: boolean; secretConfigured: boolean }
  chevereto: { baseUrl: string; keyConfigured: boolean }
  filterRules: string[]; deliveries: Delivery[]; pushStatus: string
}

class APIError extends Error {
  constructor(public status: number, message: string) { super(message) }
}

async function api<T = void>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: init?.body ? { "Content-Type": "application/json", ...init.headers } : init?.headers,
  })
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: response.statusText })) as { error?: string }
    throw new APIError(response.status, body.error || response.statusText)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

function TelegramMark({ className }: { className?: string }) {
  return <span className={cn("grid size-9 shrink-0 place-items-center rounded-full bg-[linear-gradient(180deg,#2aabee,#229ed9)] text-white", className)}>
    <svg viewBox="0 0 24 24" className="size-5" aria-hidden="true"><path fill="currentColor" d="M21.7 3.3 18.5 20c-.24 1.18-.88 1.47-1.78.92l-4.88-3.6-2.36 2.27c-.26.26-.48.48-.98.48l.35-4.97 9.04-8.17c.39-.35-.09-.55-.61-.2L6.1 13.77l-4.81-1.5c-1.05-.33-1.07-1.05.22-1.56L20.3 3.47c.87-.32 1.63.2 1.4-.17Z" /></svg>
  </span>
}

function useTheme() {
  const [dark, setDark] = useState(() => localStorage.getItem("theme") === "dark" || (!localStorage.getItem("theme") && matchMedia("(prefers-color-scheme: dark)").matches))
  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark)
    localStorage.setItem("theme", dark ? "dark" : "light")
  }, [dark])
  return [dark, setDark] as const
}

export default function App() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [authenticated, setAuthenticated] = useState<boolean | null>(null)
  const [busy, setBusy] = useState(false)
  const [toast, setToast] = useState("")
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [mobileTab, setMobileTab] = useState<"telegram" | "wechat">("telegram")
  const [dark, setDark] = useTheme()

  const load = useCallback(async (quiet = false) => {
    try {
      const data = await api<Dashboard>("/api/dashboard")
      setDashboard(data)
      setAuthenticated(true)
    } catch (error) {
      if (error instanceof APIError && error.status === 401) {
        setAuthenticated(false)
        setDashboard(null)
      } else if (!quiet) setToast(error instanceof Error ? error.message : "加载失败")
    }
  }, [])

  useEffect(() => { void load() }, [load])
  useEffect(() => {
    if (!authenticated) return
    const timer = window.setInterval(() => void load(true), dashboard?.telegram.status === "connected" ? 5000 : 2000)
    return () => clearInterval(timer)
  }, [authenticated, dashboard?.telegram.status, load])
  useEffect(() => {
    if (!toast) return
    const timer = window.setTimeout(() => setToast(""), 3500)
    return () => clearTimeout(timer)
  }, [toast])

  async function login(username: string, password: string) {
    setBusy(true)
    try {
      await api("/api/login", { method: "POST", body: JSON.stringify({ username, password }) })
      await load()
    } finally { setBusy(false) }
  }
  async function logout() {
    try {
      await api("/api/logout", { method: "POST" })
      setAuthenticated(false)
      setDashboard(null)
    } catch (error) { setToast(error instanceof Error ? error.message : "退出失败") }
  }

  if (authenticated === null) return <LoadingScreen />
  if (!authenticated || !dashboard) return <LoginScreen busy={busy} onLogin={login} dark={dark} setDark={setDark} />

  return <div className="flex h-dvh min-h-[520px] flex-col overflow-hidden bg-[var(--page)] text-foreground md:gap-2.5 md:p-2.5">
    <header className="flex h-14 shrink-0 items-center border-b border-border bg-surface px-3 sm:px-5 md:rounded-[14px] md:border-b-0 md:px-4 md:shadow-[var(--card-shadow)]">
      <TelegramMark className="size-8" />
      <div className="ml-3 min-w-0 truncate text-[15px] font-semibold">Telegram → 微信</div>
      <div className="ml-auto flex items-center gap-1">
        <StatusPill connected={dashboard.telegram.status === "connected"} label={dashboard.telegram.status === "connected" ? "Telegram 在线" : "等待登录"} />
        <Button variant="ghost" size="icon" onClick={() => setDark(!dark)} aria-label={dark ? "切换到亮色" : "切换到暗色"}>{dark ? <Sun className="size-4" /> : <Moon className="size-4" />}</Button>
        <Button variant="ghost" size="icon" onClick={() => setSettingsOpen(true)} aria-label="设置"><Settings className="size-4" /></Button>
        <Button variant="ghost" size="icon" onClick={() => void logout()} aria-label="退出控制台"><LogOut className="size-4" /></Button>
      </div>
    </header>

    <nav className="shrink-0 bg-surface px-3 pb-2.5 md:hidden">
      <div className="grid grid-cols-2 gap-1 rounded-[10px] bg-muted p-1">
        <button className={cn("rounded-[7px] py-1.5 text-[13px] font-medium transition-colors", mobileTab === "telegram" ? "bg-surface text-[var(--telegram-link)] shadow-[0_1px_2px_rgba(16,35,47,.16)]" : "text-muted-foreground")} onClick={() => setMobileTab("telegram")}>Telegram</button>
        <button className={cn("rounded-[7px] py-1.5 text-[13px] font-medium transition-colors", mobileTab === "wechat" ? "bg-surface text-[var(--wechat-accent)] shadow-[0_1px_2px_rgba(16,35,47,.16)]" : "text-muted-foreground")} onClick={() => setMobileTab("wechat")}>微信推送</button>
      </div>
    </nav>

    <main className="grid min-h-0 flex-1 md:grid-cols-2 md:gap-2.5">
      <TelegramPanel dashboard={dashboard} visible={mobileTab === "telegram"} setDashboard={setDashboard} reload={load} notify={setToast} />
      <WechatPanel dashboard={dashboard} visible={mobileTab === "wechat"} reload={load} notify={setToast} />
    </main>

    <SettingsPanel open={settingsOpen} setOpen={setSettingsOpen} dashboard={dashboard} reload={load} notify={setToast} />
    {toast && <div role="status" className="fixed bottom-6 left-1/2 z-[80] -translate-x-1/2 rounded-full bg-[#000000d9] px-4 py-2 text-sm text-white shadow-[var(--panel-shadow)] backdrop-blur">{toast}</div>}
  </div>
}

function LoadingScreen() {
  return <div className="grid min-h-dvh place-items-center bg-[var(--page)]"><div className="flex items-center gap-3 text-muted-foreground"><LoaderCircle className="size-5 animate-spin text-[var(--telegram-link)]" /> 正在连接控制台…</div></div>
}

function LoginScreen({ busy, onLogin, dark, setDark }: { busy: boolean; onLogin: (u: string, p: string) => Promise<void>; dark: boolean; setDark: (v: boolean) => void }) {
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  async function submit(event: FormEvent) {
    event.preventDefault(); setError("")
    try { await onLogin(username, password) } catch (reason) { setError(reason instanceof Error ? reason.message : "登录失败") }
  }
  return <div className="relative grid min-h-dvh place-items-center bg-[var(--page)] px-5">
    <Button className="absolute right-5 top-5" variant="ghost" size="icon" onClick={() => setDark(!dark)} aria-label="切换主题">{dark ? <Sun className="size-4" /> : <Moon className="size-4" />}</Button>
    <form onSubmit={submit} className="w-full max-w-[380px] rounded-[16px] bg-surface p-7 text-center shadow-[var(--panel-shadow)] sm:p-9">
      <TelegramMark className="mx-auto mb-5 size-20 [&_svg]:size-11" />
      <h1 className="text-[22px] font-semibold">Telegram Forwarder</h1>
      <p className="mb-7 mt-2 text-[15px] leading-6 text-muted-foreground">登录后管理 Telegram 到微信的消息转发。</p>
      <div className="text-left">
        <label className="mb-1.5 block text-xs font-semibold text-muted-foreground" htmlFor="username">用户名</label>
        <Input id="username" autoComplete="username" value={username} onChange={(e) => setUsername(e.target.value)} required autoFocus />
        <label className="mb-1.5 mt-4 block text-xs font-semibold text-muted-foreground" htmlFor="password">密码</label>
        <Input id="password" type="password" autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} required />
      </div>
      {error && <p className="mt-3 flex items-center gap-2 text-left text-sm text-red-600 dark:text-red-400"><CircleAlert className="size-4 shrink-0" />{error}</p>}
      <Button className="mt-6 w-full rounded-full" disabled={busy}>{busy && <LoaderCircle className="size-4 animate-spin" />}{busy ? "登录中…" : "登录"}</Button>
    </form>
  </div>
}

function StatusPill({ connected, label }: { connected: boolean; label: string }) {
  return <span className="mr-1 hidden items-center gap-1.5 rounded-full bg-muted px-3 py-1 text-[11px] text-muted-foreground sm:inline-flex"><span className={cn("size-1.5 rounded-full", connected ? "bg-[#4dcd5e]" : "bg-amber-500")} />{label}</span>
}

function TelegramPanel({ dashboard, visible, setDashboard, reload, notify }: { dashboard: Dashboard; visible: boolean; setDashboard: Dispatch<SetStateAction<Dashboard | null>>; reload: (quiet?: boolean) => Promise<void>; notify: (s: string) => void }) {
  const [search, setSearch] = useState("")
  const [selectedOnly, setSelectedOnly] = useState(false)
  const [archiveOpen, setArchiveOpen] = useState(false)
  const telegram = dashboard.telegram
  const dialogs = useMemo(() => dashboard.dialogs.filter((item) => {
    const matches = [item.title, item.subtitle, item.username].join(" ").toLowerCase().includes(search.toLowerCase())
    return matches && (!selectedOnly || item.selected)
  }), [dashboard.dialogs, search, selectedOnly])
  const active = dialogs.filter((item) => !item.archived)
  const archived = dialogs.filter((item) => item.archived)
  const showArchived = archiveOpen || search.trim() !== ""

  async function patchDialog(item: DialogItem, change: Partial<Pick<DialogItem, "selected" | "adFilter">>) {
    setDashboard((current) => current ? { ...current, dialogs: current.dialogs.map((d) => d.peerKey === item.peerKey ? { ...d, ...change } : d) } : current)
    try {
      await api(`/api/dialogs/${encodeURIComponent(item.peerKey)}`, { method: "PATCH", body: JSON.stringify(change) })
    } catch (error) {
      notify(error instanceof Error ? error.message : "更新失败")
      await reload(true)
    }
  }

  async function refreshTelegram() {
    try { await api("/api/telegram/refresh", { method: "POST" }); await reload(true); notify("会话列表已刷新") }
    catch (error) { notify(error instanceof Error ? error.message : "刷新失败") }
  }

  async function logoutTelegram() {
    try { await api("/api/telegram/logout", { method: "POST" }); await reload(true) }
    catch (error) { notify(error instanceof Error ? error.message : "退出 Telegram 失败") }
  }

  return <section className={cn("min-h-0 overflow-hidden bg-surface md:flex md:flex-col md:rounded-[14px] md:shadow-[var(--card-shadow)]", visible ? "flex flex-col" : "hidden")} aria-label="Telegram 会话">
    <div className="flex h-16 shrink-0 items-center border-b border-border px-4">
      <div className="grid size-10 place-items-center rounded-full bg-[linear-gradient(180deg,#2aabee,#229ed9)] text-white"><Send className="size-5" /></div>
      <div className="ml-3 min-w-0"><h2 className="truncate text-[15px] font-semibold">Telegram</h2><p className="truncate text-xs text-muted-foreground">{telegram.account || statusText(telegram.status)}</p></div>
      <div className="ml-auto flex gap-1">
        {telegram.status === "connected" && <Button variant="ghost" size="icon" onClick={() => void refreshTelegram()} aria-label="刷新会话"><RefreshCw className="size-4" /></Button>}
        {telegram.status === "connected" && <Button variant="ghost" size="icon" onClick={() => void logoutTelegram()} aria-label="退出 Telegram"><WifiOff className="size-4" /></Button>}
      </div>
    </div>

    {telegram.status !== "connected" ? <TelegramLogin state={telegram} reload={reload} notify={notify} /> : <>
      <div className="shrink-0 px-3 pb-2 pt-3">
        <div className="relative"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><Input className="h-9 rounded-[10px] pl-9" placeholder="搜索对话" value={search} onChange={(e) => setSearch(e.target.value)} /></div>
        <div className="mt-2.5 flex items-center justify-between px-1 text-xs"><span className="text-muted-foreground">{dashboard.dialogs.filter((d) => d.selected).length} 个来源已启用</span><label className="flex items-center gap-2 text-muted-foreground"><Switch checked={selectedOnly} onCheckedChange={setSelectedOnly} />只看已选</label></div>
      </div>
      <div className="scrollbar min-h-0 flex-1 space-y-0.5 overflow-y-auto px-2 pb-3">
        {active.map((item) => <DialogRow key={item.peerKey} item={item} patch={patchDialog} />)}
        {archived.length > 0 && <>
          <button type="button" onClick={() => setArchiveOpen(!archiveOpen)} aria-expanded={showArchived} className="flex min-h-[60px] w-full items-center gap-2.5 rounded-[10px] px-2.5 py-2 text-left transition-colors hover:bg-muted">
            <span className="grid size-11 shrink-0 place-items-center rounded-full bg-muted text-muted-foreground"><Archive className="size-5" /></span>
            <span className="flex-1 text-[15px] font-semibold">已归档的聊天</span>
            <span className="rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">{archived.length}</span>
            <ChevronRight className={cn("size-4 shrink-0 text-muted-foreground transition-transform", showArchived && "rotate-90")} />
          </button>
          {showArchived && archived.map((item) => <DialogRow key={item.peerKey} item={item} patch={patchDialog} />)}
        </>}
        {!dialogs.length && <div className="grid h-48 place-items-center text-sm text-muted-foreground">没有匹配的对话</div>}
      </div>
    </>}
  </section>
}

function TelegramLogin({ state, reload, notify }: { state: TelegramState; reload: (quiet?: boolean) => Promise<void>; notify: (s: string) => void }) {
  const [password, setPassword] = useState("")
  if (state.status === "qr" && state.qrCode) return <div className="scrollbar grid min-h-0 flex-1 place-items-center overflow-auto p-8 text-center"><div>
    <img className="mx-auto size-56 rounded-[16px] bg-white p-3 shadow-[var(--panel-shadow)]" src={state.qrCode} alt="Telegram 登录二维码" />
    <h3 className="mt-5 text-[20px] font-semibold">扫描二维码登录</h3><p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-muted-foreground">打开 Telegram → 设置 → 设备 → 链接桌面设备。二维码会自动刷新。</p>
  </div></div>
  if (state.status === "password") return <form className="m-auto w-full max-w-sm p-8" onSubmit={async (event) => {
    event.preventDefault()
    try { await api("/api/telegram/password", { method: "POST", body: JSON.stringify({ password }) }); setPassword(""); await reload(true) } catch (error) { notify(error instanceof Error ? error.message : "验证失败") }
  }}>
    <LockKeyhole className="mb-4 size-9 text-[var(--telegram-link)]" /><h3 className="text-xl font-semibold">两步验证</h3><p className="mb-5 mt-2 text-sm text-muted-foreground">该 Telegram 账号启用了云密码，请输入后继续。</p><Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoFocus required /><Button className="mt-4 w-full">验证</Button>{state.error && <p className="mt-3 text-sm text-red-500">{state.error}</p>}
  </form>
  return <div className="grid min-h-0 flex-1 place-items-center p-8 text-center"><div><LoaderCircle className="mx-auto size-8 animate-spin text-[var(--telegram-link)]" /><h3 className="mt-4 text-lg text-[var(--telegram-link)]">{statusText(state.status)}</h3>{state.error && <p className="mt-2 max-w-sm text-sm text-red-500">{state.error}</p>}</div></div>
}

function DialogRow({ item, patch }: { item: DialogItem; patch: (item: DialogItem, change: Partial<Pick<DialogItem, "selected" | "adFilter">>) => Promise<void> }) {
  const initials = item.title.trim().slice(0, 2).toUpperCase() || "TG"
  return <>
    <div className={cn("group flex min-h-[60px] items-center rounded-[10px] px-2.5 py-2 transition-colors", item.selected ? "bg-[var(--selected)] text-white" : "hover:bg-muted")}>
      <div className="grid size-11 shrink-0 place-items-center rounded-full text-[15px] font-semibold text-white [text-shadow:0_1px_2px_rgba(0,0,0,.25)]" style={{ background: avatarColor(item.peerKey) }}>{initials}</div>
      <div className="ml-2.5 min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-[15px] font-semibold">{item.title}</span>
          {item.subtitle && <span className={cn("shrink-0 rounded-full px-1.5 py-px text-[10px] font-medium", item.selected ? "bg-white/25 text-white" : "bg-muted text-muted-foreground")}>{item.subtitle}</span>}
          {item.pinned && <Pin className={cn("size-3.5 shrink-0", item.selected ? "text-white/80" : "text-muted-foreground")} aria-label="已置顶" />}
        </div>
        <div className={cn("mt-0.5 truncate text-[13px]", item.selected ? "text-white/75" : "text-muted-foreground")}>
          {item.lastSender && <span className={item.selected ? "text-white" : "text-foreground/75"}>{item.lastSender}: </span>}
          {item.lastText || (item.lastSender ? "" : item.username)}
        </div>
      </div>
      {item.selectable && <div className="ml-2 flex items-center gap-2">
        {item.selected && <button type="button" className={cn("grid size-8 place-items-center rounded-full transition-colors hover:bg-white/20", item.adFilter ? "text-white" : "text-white/55")} onClick={() => void patch(item, { adFilter: !item.adFilter })} aria-label={item.adFilter ? "关闭广告过滤" : "开启广告过滤"} title={item.adFilter ? "广告过滤已开启" : "广告过滤已关闭"}><Filter className="size-4" /></button>}
        <Checkbox className={item.selected ? "border-white/70 focus-visible:ring-white focus-visible:ring-offset-[var(--selected)] data-[state=checked]:border-white data-[state=checked]:bg-white data-[state=checked]:text-[var(--selected)]" : undefined} checked={item.selected} onCheckedChange={(checked) => void patch(item, { selected: checked === true })} aria-label={`转发 ${item.title}`} />
      </div>}
    </div>
  </>
}

function WechatPanel({ dashboard, visible, reload, notify }: { dashboard: Dashboard; visible: boolean; reload: (quiet?: boolean) => Promise<void>; notify: (s: string) => void }) {
  const ready = dashboard.pushStatus === "ready"
  const [draft, setDraft] = useState("")
  const [sending, setSending] = useState(false)

  async function send(event: FormEvent) {
    event.preventDefault()
    const content = draft.trim()
    if (!content || sending) return
    setSending(true)
    try {
      await api("/api/push/message", { method: "POST", body: JSON.stringify({ content }) })
      setDraft("")
      await reload(true)
    } catch (error) {
      notify(error instanceof Error ? error.message : "发送失败")
    } finally { setSending(false) }
  }

  return <section className={cn("wechat-chat relative min-h-0 overflow-hidden md:flex md:flex-col md:rounded-[14px] md:shadow-[var(--card-shadow)]", visible ? "flex flex-col" : "hidden")} aria-label="微信推送">
    <div className="glass-card absolute inset-x-2 top-2 z-10 flex h-14 items-center px-3">
      <div className="grid size-9 place-items-center rounded-full bg-[linear-gradient(180deg,#5ad477,#07c160)] text-white"><MessageCircle className="size-[18px]" /></div>
      <div className="ml-2.5 min-w-0"><h2 className="truncate text-[15px] font-semibold">PushPlus 微信</h2><p className="truncate text-xs text-muted-foreground">{ready ? "已就绪" : dashboard.pushStatus.startsWith("limited") ? "额度受限" : "等待配置"}</p></div>
      <span className={cn("ml-auto flex shrink-0 items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-xs", ready ? "text-[var(--wechat-accent)]" : "text-muted-foreground")}>{ready ? <Wifi className="size-3.5" /> : <WifiOff className="size-3.5" />}{ready ? "在线" : "未连接"}</span>
    </div>
    <div className="scrollbar flex min-h-0 flex-1 flex-col-reverse gap-2 overflow-y-auto px-3 pb-1 pt-[74px] sm:px-6">
      {dashboard.deliveries.length ? dashboard.deliveries.map((item, index) => <DeliveryBubble key={`${item.shortCode || item.title}-${index}`} item={item} />) : <div className="m-auto rounded-[14px] bg-black/25 px-6 py-5 text-center text-white backdrop-blur-sm"><MessageCircle className="mx-auto mb-3 size-10 opacity-70" /><p className="text-sm">还没有推送记录</p><p className="mt-1 text-xs opacity-80">配置 PushPlus 后，新消息会显示在这里</p></div>}
    </div>
    <form onSubmit={send} className="glass-card z-10 m-2 flex h-14 shrink-0 items-center gap-2 px-2.5">
      <Input className="h-10 flex-1 rounded-full" placeholder="发送消息到微信" value={draft} onChange={(e) => setDraft(e.target.value)} aria-label="发送消息到微信" maxLength={2000} />
      <Button size="icon" disabled={sending || !draft.trim()} aria-label="发送">{sending ? <LoaderCircle className="size-4 animate-spin" /> : <Send className="size-4" />}</Button>
    </form>
  </section>
}

function DeliveryBubble({ item }: { item: Delivery }) {
  const content = item.content ? plainContent(item.content) : "已通过 PushPlus 推送"
  return <div className="ml-auto max-w-[88%] sm:max-w-[72%]">
    <div className="wechat-bubble relative min-w-[152px] px-2.5 py-1.5">
      <div className="text-[13px] font-semibold text-[var(--wechat-accent)]">{item.title}</div>
      <div className="mt-0.5 whitespace-pre-wrap break-words text-[14px] leading-[19px]">{truncateText(content, 420)}</div>
      {item.error && <div className="mt-1 max-w-[240px] truncate text-[11px] text-red-700 dark:text-red-300" title={item.error}>{item.error}</div>}
      <div className="mt-0.5 flex items-center justify-end gap-1 text-[11px] text-[var(--wechat-bubble-meta)]">{deliveryStatus(item.status)}<span>{item.time}</span>{item.status === "failed" ? <CircleAlert className="size-3" /> : <Check className="size-3" />}</div>
    </div>
  </div>
}

function SettingsPanel({ open, setOpen, dashboard, reload, notify }: { open: boolean; setOpen: (v: boolean) => void; dashboard: Dashboard; reload: (quiet?: boolean) => Promise<void>; notify: (s: string) => void }) {
  const [token, setToken] = useState("")
  const [secret, setSecret] = useState("")
  const [topic, setTopic] = useState(dashboard.pushplus.topic || "")
  const [baseURL, setBaseURL] = useState(dashboard.chevereto.baseUrl || "")
  const [imageKey, setImageKey] = useState("")
  const [rules, setRules] = useState(dashboard.filterRules.join("\n"))
  const [saving, setSaving] = useState(false)
  useEffect(() => {
    if (!open) return
    setToken(""); setSecret(""); setImageKey(""); setTopic(dashboard.pushplus.topic || ""); setBaseURL(dashboard.chevereto.baseUrl || ""); setRules(dashboard.filterRules.join("\n"))
  }, [open])

  async function saveAll(showMessage = true) {
    setSaving(true)
    try {
      await Promise.all([
        api("/api/settings/pushplus", { method: "PUT", body: JSON.stringify({ token, secretKey: secret, topic }) }),
        api("/api/settings/chevereto", { method: "PUT", body: JSON.stringify({ baseUrl: baseURL, apiKey: imageKey }) }),
        api("/api/settings/filters", { method: "PUT", body: JSON.stringify({ rules: rules.split("\n") }) }),
      ])
      setToken(""); setSecret(""); setImageKey(""); await reload(true)
      if (showMessage) notify("设置已保存")
      return true
    } catch (error) {
      notify(error instanceof Error ? error.message : "保存失败")
      return false
    } finally { setSaving(false) }
  }
  async function test(kind: "push" | "image") {
    try {
      if (!await saveAll(false)) return
      const route = kind === "push" ? "pushplus" : "chevereto"
      const result = await api<Record<string, string>>(`/api/settings/${route}/test`, { method: "POST" })
      notify(kind === "push" ? `测试推送已提交 · ${result.shortCode || "OK"}` : "测试图片上传成功")
    } catch (error) { notify(error instanceof Error ? error.message : "测试失败") }
  }
  async function clear(kind: "push" | "image") {
    if (!confirm("确定清除已保存的凭据吗？")) return
    try {
      if (kind === "push") await api("/api/settings/pushplus", { method: "PUT", body: JSON.stringify({ topic: "", clearToken: true, clearSecret: true }) })
      else await api("/api/settings/chevereto", { method: "PUT", body: JSON.stringify({ baseUrl: "", clearKey: true }) })
      await reload(true); notify("凭据已清除")
    } catch (error) { notify(error instanceof Error ? error.message : "清除失败") }
  }

  return <Dialog open={open} onOpenChange={setOpen}><DialogContent>
    <DialogHeader><DialogTitle className="text-lg font-semibold">转发设置</DialogTitle><DialogDescription className="mt-1 text-sm text-muted-foreground">密钥加密保存在 ./data/data.db，保存后不会再次回传到浏览器。</DialogDescription></DialogHeader>
    <div className="space-y-2.5 bg-[var(--page)] px-3 py-3 sm:px-4 sm:py-4">
      <SettingsSection icon={<MessageCircle className="size-4" />} title="PushPlus 微信" description="所有已选 Telegram 会话共用这一个目标。">
        <Field label="用户 Token" hint={dashboard.pushplus.tokenConfigured ? "已配置；留空不修改" : "必填"}><Input type="password" value={token} onChange={(e) => setToken(e.target.value)} placeholder={dashboard.pushplus.tokenConfigured ? "••••••••••••" : "PushPlus Token"} autoComplete="off" /></Field>
        <Field label="SecretKey" hint={dashboard.pushplus.secretConfigured ? "已配置；用于历史与投递状态" : "开放接口必填"}><Input type="password" value={secret} onChange={(e) => setSecret(e.target.value)} placeholder={dashboard.pushplus.secretConfigured ? "••••••••••••" : "PushPlus SecretKey"} autoComplete="off" /></Field>
        <Field label="Topic" hint="可选；为空时只推送给自己"><Input value={topic} onChange={(e) => setTopic(e.target.value)} placeholder="群组编码" /></Field>
        <div className="flex flex-wrap gap-2"><Button variant="outline" size="sm" onClick={() => void test("push")}>发送测试消息</Button>{dashboard.pushplus.tokenConfigured && <Button variant="danger" size="sm" onClick={() => void clear("push")}>清除凭据</Button>}</div>
      </SettingsSection>
      <SettingsSection icon={<Image className="size-4" />} title="Chevereto 图床" description="Telegram 图片与静态贴纸上传后嵌入 PushPlus HTML 消息。">
        <Field label="API Base URL" hint="站点根地址、/api/1 或完整上传地址"><Input type="url" value={baseURL} onChange={(e) => setBaseURL(e.target.value)} placeholder="https://img.example.com" /></Field>
        <Field label="API Key" hint={dashboard.chevereto.keyConfigured ? "已配置；留空不修改" : "必填"}><Input type="password" value={imageKey} onChange={(e) => setImageKey(e.target.value)} placeholder={dashboard.chevereto.keyConfigured ? "••••••••••••" : "Chevereto API Key"} autoComplete="off" /></Field>
        <div className="flex flex-wrap gap-2"><Button variant="outline" size="sm" onClick={() => void test("image")}>测试上传</Button>{dashboard.chevereto.keyConfigured && <Button variant="danger" size="sm" onClick={() => void clear("image")}>清除凭据</Button>}</div>
      </SettingsSection>
      <SettingsSection icon={<ShieldCheck className="size-4" />} title="广告过滤规则" description="仅对开启广告过滤的会话生效；检查发送者、文本和图片说明。">
        <Field label="每行一条规则" hint="普通行按关键词；正则以 re: 开头"><Textarea rows={7} value={rules} onChange={(e) => setRules(e.target.value)} placeholder={"推广\n加群\nre:(?i)limited offer|promo code"} /></Field>
      </SettingsSection>
    </div>
    <div className="sticky bottom-0 flex items-center justify-end gap-2 border-t border-border bg-background px-6 py-4"><Button variant="ghost" onClick={() => setOpen(false)}>取消</Button><Button disabled={saving} onClick={() => void saveAll()}>{saving && <LoaderCircle className="size-4 animate-spin" />}保存设置</Button></div>
  </DialogContent></Dialog>
}

function SettingsSection({ icon, title, description, children }: { icon: ReactNode; title: string; description: string; children: ReactNode }) {
  return <section className="rounded-[14px] bg-surface p-4 shadow-[var(--card-shadow)] sm:p-5"><div className="mb-4 flex items-start gap-3"><div className="grid size-8 shrink-0 place-items-center rounded-full bg-muted text-[var(--telegram-link)]">{icon}</div><div><h3 className="font-semibold">{title}</h3><p className="mt-0.5 text-xs leading-5 text-muted-foreground">{description}</p></div></div><div className="space-y-4">{children}</div></section>
}
function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return <label className="block"><span className="mb-1.5 flex items-center justify-between gap-4 text-xs font-semibold"><span>{label}</span>{hint && <span className="text-right font-normal text-muted-foreground">{hint}</span>}</span>{children}</label>
}
function statusText(status: string) { return ({ starting: "正在启动", connecting: "正在连接", qr: "等待扫码", password: "需要两步验证", error: "连接异常", disconnected: "连接已断开" } as Record<string, string>)[status] || status }
function deliveryStatus(status: string) { return ({ queued: "排队中", accepted: "已受理", sent: "已送达", failed: "失败", expired: "已过期", history: "历史" } as Record<string, string>)[status] || status }
function plainContent(value: string) { return value.replace(/<br\s*\/?\s*>/gi, "\n").replace(/<img[^>]*>/gi, "[图片]").replace(/<[^>]+>/g, "").replaceAll("&lt;", "<").replaceAll("&gt;", ">").replaceAll("&amp;", "&") }
function truncateText(value: string, max: number) { return value.length > max ? value.slice(0, max - 1) + "…" : value }
// Telegram's seven avatar gradients, picked by a stable hash of the peer key.
function avatarColor(key: string) {
  const gradients = [["#ff885e", "#ff516a"], ["#ffcd6a", "#ffa85c"], ["#82b1ff", "#665fff"], ["#a0de7e", "#54cb68"], ["#53edd6", "#28c9b7"], ["#72d5fd", "#2a9ef1"], ["#e0a2f3", "#d669ed"]]
  let hash = 0
  for (const char of key) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  const [from, to] = gradients[hash % gradients.length]
  return `linear-gradient(180deg, ${from}, ${to})`
}
