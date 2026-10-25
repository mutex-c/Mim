/**
 * Deterministic high-contrast avatar color generator & status ring utilities.
 * Ensures avatars are always clearly visible and never wash out against any background.
 */

export interface AvatarColor {
  bg: string;
  text: string;
}

const AVATAR_PALETTES: AvatarColor[] = [
  { bg: '#1e5243', text: '#ffffff' }, // Alpine Pine
  { bg: '#244855', text: '#ffffff' }, // Baltic Petrol
  { bg: '#8a4b2d', text: '#ffffff' }, // Warm Terracotta
  { bg: '#543d63', text: '#ffffff' }, // Deep Mulberry
  { bg: '#6e4f32', text: '#ffffff' }, // Warm Bronze Teak
  { bg: '#2b5858', text: '#ffffff' }, // Nordic Deep Spruce
  { bg: '#36454f', text: '#ffffff' }, // Monolith Slate
  { bg: '#7b3f46', text: '#ffffff' }, // Smoked Mahogany
];

const GROUP_PALETTE: AvatarColor = {
  bg: '#7c5535',
  text: '#ffffff',
};

/**
 * Returns a deterministic, rich, high-contrast background and text color
 * for any user or group ID/name.
 */
export function getAvatarColor(idOrName?: string | number | null, isGroup = false): AvatarColor {
  if (isGroup) {
    return GROUP_PALETTE;
  }
  const str = String(idOrName ?? 'User');
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  const index = Math.abs(hash) % AVATAR_PALETTES.length;
  return AVATAR_PALETTES[index];
}

export type AvatarStatus = 'online' | 'offline' | 'busy' | 'group';

/**
 * Returns Tailwind ring classes representing different contact statuses.
 * - Online: Emerald ring (🟢)
 * - Busy/Away: Amber ring (🟡)
 * - Offline: Sleek subtle border (⚪)
 * - Group: Distinctive bronze/group badge ring (👥)
 */
export function getStatusRingClass(status: AvatarStatus | boolean | undefined): string {
  if (status === 'group') {
    return 'ring-2 ring-amber-700/50 dark:ring-amber-400/50 ring-offset-2 ring-offset-white dark:ring-offset-[#1b201e]';
  }
  if (status === true || status === 'online') {
    return 'ring-2 ring-emerald-500 ring-offset-2 ring-offset-white dark:ring-offset-[#1b201e]';
  }
  if (status === 'busy') {
    return 'ring-2 ring-amber-500 ring-offset-2 ring-offset-white dark:ring-offset-[#1b201e]';
  }
  // Offline
  return 'ring-1.5 ring-black/15 dark:ring-white/20 ring-offset-1 ring-offset-white dark:ring-offset-[#1b201e] opacity-90';
}
