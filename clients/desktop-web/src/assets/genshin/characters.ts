export interface GenshinCharacter {
  id: string;
  name: string;
  title: string;
  element: 'Anemo' | 'Geo' | 'Electro' | 'Dendro' | 'Hydro' | 'Pyro' | 'Cryo' | 'Universal';
  elementName: string;
  elementColor: string;
  quote: string;
  avatarSvg: string;
}

export const GENSHIN_CHARACTERS: GenshinCharacter[] = [
  {
    id: 'paimon',
    name: '派蒙',
    title: '最好的伙伴 · 应急食品',
    element: 'Universal',
    elementName: '天权星宿',
    elementColor: '#e5b044',
    quote: '前面的区域，以后再来探索吧！',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="paimonBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%23fff9eb"/>
          <stop offset="100%" stop-color="%23f7e5b8"/>
        </radialGradient>
        <linearGradient id="goldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="%23fcd34d"/>
          <stop offset="100%" stop-color="%23d97706"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23paimonBg)" stroke="%23e5b044" stroke-width="3"/>
      <!-- Halo/Crown -->
      <polygon points="50,14 54,24 64,24 56,30 59,40 50,34 41,40 44,30 36,24 46,24" fill="url(%23goldGrad)"/>
      <!-- Hair Base -->
      <path d="M26,50 C26,30 74,30 74,50 C74,70 65,75 50,75 C35,75 26,70 26,50 Z" fill="%23ffffff" stroke="%23e5e7eb" stroke-width="1.5"/>
      <!-- Hair Bangs -->
      <path d="M36,36 C42,48 48,48 50,38 C52,48 58,48 64,36" fill="none" stroke="%23d1d5db" stroke-width="2.5" stroke-linecap="round"/>
      <!-- Eyes -->
      <ellipse cx="41" cy="52" rx="4" ry="6" fill="%231e3a8a"/>
      <ellipse cx="59" cy="52" rx="4" ry="6" fill="%231e3a8a"/>
      <circle cx="42" cy="50" r="1.5" fill="%23ffffff"/>
      <circle cx="60" cy="50" r="1.5" fill="%23ffffff"/>
      <!-- Blush -->
      <ellipse cx="34" cy="58" rx="4" ry="2" fill="%23fca5a5" opacity="0.6"/>
      <ellipse cx="66" cy="58" rx="4" ry="2" fill="%23fca5a5" opacity="0.6"/>
      <!-- Smile -->
      <path d="M46,60 Q50,64 54,60" fill="none" stroke="%23e11d48" stroke-width="2" stroke-linecap="round"/>
      <!-- Primogem Sparkle -->
      <polygon points="50,74 53,80 59,80 54,84 56,90 50,86 44,90 46,84 41,80 47,80" fill="%2338bdf8"/>
    </svg>`,
  },
  {
    id: 'zhongli',
    name: '钟离',
    title: '往生堂客卿 · 摩拉克斯',
    element: 'Geo',
    elementName: '岩元素',
    elementColor: '#cfa858',
    quote: '欲买桂花同载酒，只可惜故人，何日再见呢。',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="geoBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%23261f18"/>
          <stop offset="100%" stop-color="%23120e0b"/>
        </radialGradient>
        <linearGradient id="amberGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="%23fef08a"/>
          <stop offset="100%" stop-color="%23ca8a04"/>
        </linearGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23geoBg)" stroke="%23cfa858" stroke-width="3"/>
      <!-- Hair Dark & Amber Tip -->
      <path d="M30,30 C35,16 65,16 70,30 C76,45 74,68 68,82 C55,88 45,88 32,82 C26,68 24,45 30,30 Z" fill="%231c1917"/>
      <!-- Face -->
      <path d="M35,38 C35,58 40,68 50,72 C60,68 65,58 65,38 Z" fill="%23fef3c7"/>
      <!-- Sharp Golden Eyes with Red Eyeliner -->
      <path d="M38,48 L46,47 L44,52 Z" fill="%23dc2626"/>
      <ellipse cx="42" cy="50" rx="3.5" ry="4.5" fill="url(%23amberGrad)"/>
      <rect x="41.5" y="47" width="1" height="6" fill="%2378350f"/>
      <path d="M62,48 L54,47 L56,52 Z" fill="%23dc2626"/>
      <ellipse cx="58" cy="50" rx="3.5" ry="4.5" fill="url(%23amberGrad)"/>
      <rect x="57.5" y="47" width="1" height="6" fill="%2378350f"/>
      <!-- Gentle calm mouth -->
      <line x1="48" y1="62" x2="52" y2="62" stroke="%2392400e" stroke-width="1.5" stroke-linecap="round"/>
      <!-- Geo Symbol on Bottom -->
      <rect x="44" y="78" width="12" height="12" transform="rotate(45 50 84)" fill="%23eab308" stroke="%2378350f" stroke-width="1.5"/>
    </svg>`,
  },
  {
    id: 'furina',
    name: '芙宁娜',
    title: '戏剧与正义 · 枫丹伶人',
    element: 'Hydro',
    elementName: '水元素',
    elementColor: '#4ba6f5',
    quote: '喝彩吧！沉醉吧！属于我的盛大剧目！',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="hydroBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%231e293b"/>
          <stop offset="100%" stop-color="%230f172a"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23hydroBg)" stroke="%2338bdf8" stroke-width="3"/>
      <!-- Blue Top Hat with Ribbon -->
      <path d="M28,24 L72,24 L68,6 L32,6 Z" fill="%231e3a8a" stroke="%2338bdf8" stroke-width="1.5"/>
      <ellipse cx="50" cy="24" rx="28" ry="5" fill="%230284c7"/>
      <!-- Soft White Blue Hair -->
      <path d="M25,40 C28,26 72,26 75,40 C80,60 72,75 50,78 C28,75 20,60 25,40 Z" fill="%23f0f9ff"/>
      <!-- Face -->
      <path d="M34,42 C34,60 40,68 50,70 C60,68 66,60 66,42 Z" fill="%23fdf4ff"/>
      <!-- Dual Heterochromia & Droplet Eyes -->
      <ellipse cx="42" cy="52" rx="4" ry="5" fill="%2338bdf8"/>
      <ellipse cx="58" cy="52" rx="4" ry="5" fill="%230284c7"/>
      <circle cx="43" cy="50" r="1.5" fill="%23ffffff"/>
      <circle cx="59" cy="50" r="1.5" fill="%23ffffff"/>
      <!-- Playful drama smile -->
      <path d="M46,62 Q50,66 54,62" fill="none" stroke="%230369a1" stroke-width="2" stroke-linecap="round"/>
      <!-- Water Droplet Jewel -->
      <path d="M50,78 C46,84 46,90 50,92 C54,90 54,84 50,78 Z" fill="%2338bdf8"/>
    </svg>`,
  },
  {
    id: 'nahida',
    name: '纳西妲',
    title: '小吉祥草王 · 布耶尔',
    element: 'Dendro',
    elementName: '草元素',
    elementColor: '#76c867',
    quote: '此即，智慧之殿堂！',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="dendroBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%23f0fdf4"/>
          <stop offset="100%" stop-color="%23dcfce7"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23dendroBg)" stroke="%234ade80" stroke-width="3"/>
      <!-- White & Leaf Green Hair with Ponytail -->
      <path d="M28,42 C28,24 72,24 72,42 C78,65 72,74 50,76 C28,74 22,65 28,42 Z" fill="%23ffffff"/>
      <path d="M68,36 C78,42 82,54 78,62 C74,58 72,48 68,36 Z" fill="%2386efac"/>
      <!-- Face -->
      <path d="M35,42 C35,60 40,68 50,70 C60,68 65,60 65,42 Z" fill="%23fefce8"/>
      <!-- Clover Green Eyes -->
      <ellipse cx="42" cy="52" rx="4" ry="5.5" fill="%2322c55e"/>
      <ellipse cx="58" cy="52" rx="4" ry="5.5" fill="%2322c55e"/>
      <circle cx="43" cy="50" r="1.5" fill="%23ffffff"/>
      <circle cx="59" cy="50" r="1.5" fill="%23ffffff"/>
      <!-- Gentle smile -->
      <path d="M47,62 Q50,65 53,62" fill="none" stroke="%2315803d" stroke-width="2" stroke-linecap="round"/>
      <!-- Dendro Leaf Ornament -->
      <path d="M50,14 C56,22 62,22 60,28 C54,28 48,22 50,14 Z" fill="%234ade80"/>
      <circle cx="50" cy="84" r="6" fill="%2386efac" stroke="%2316a34a" stroke-width="1.5"/>
    </svg>`,
  },
  {
    id: 'raiden',
    name: '雷电将军',
    title: '御建鸣神主尊大御所 · 巴尔泽布',
    element: 'Electro',
    elementName: '雷元素',
    elementColor: '#9d6af5',
    quote: '此身，当为稻妻之极致！',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="electroBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%232e1065"/>
          <stop offset="100%" stop-color="%2317072b"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23electroBg)" stroke="%23a855f7" stroke-width="3"/>
      <!-- Deep Purple Long Braid -->
      <path d="M28,34 C30,18 70,18 72,34 C76,52 74,74 50,82 C26,74 24,52 28,34 Z" fill="%233b0764"/>
      <!-- Hair Ornament: Golden Fan & Purple Blossom -->
      <polygon points="70,22 84,28 78,36" fill="%23fbbf24"/>
      <circle cx="72" cy="30" r="4" fill="%23c084fc"/>
      <!-- Face -->
      <path d="M35,38 C35,58 40,68 50,72 C60,68 65,58 65,38 Z" fill="%23faf5ff"/>
      <!-- Purple Eyes & Beauty Mark -->
      <ellipse cx="42" cy="50" rx="3.5" ry="4.5" fill="%23a855f7"/>
      <circle cx="43" cy="48" r="1.2" fill="%23ffffff"/>
      <ellipse cx="58" cy="50" rx="3.5" ry="4.5" fill="%23a855f7"/>
      <circle cx="59" cy="48" r="1.2" fill="%23ffffff"/>
      <circle cx="61" cy="57" r="1" fill="%23581c87"/> <!-- Beauty Mark -->
      <line x1="48" y1="63" x2="52" y2="63" stroke="%237e22ce" stroke-width="1.5" stroke-linecap="round"/>
      <!-- Electro Crest Emblem -->
      <circle cx="50" cy="86" r="6" fill="%23c084fc" stroke="%23e9d5ff" stroke-width="1.5"/>
    </svg>`,
  },
  {
    id: 'hutao',
    name: '胡桃',
    title: '往生堂第七十七代堂主',
    element: 'Pyro',
    elementName: '火元素',
    elementColor: '#e85642',
    quote: '大丘丘病了二丘丘瞧，三丘丘采药四丘丘熬～',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="pyroBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%233b1510"/>
          <stop offset="100%" stop-color="%231a0806"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23pyroBg)" stroke="%23f97316" stroke-width="3"/>
      <!-- Porkpie Hat with Red Plum Blossom -->
      <ellipse cx="50" cy="22" rx="26" ry="6" fill="%231c1917"/>
      <rect x="30" y="10" width="40" height="12" fill="%23292524"/>
      <circle cx="34" cy="20" r="4.5" fill="%23ef4444"/>
      <!-- Dark Brown Hair -->
      <path d="M26,36 C28,24 72,24 74,36 C78,56 74,74 50,78 C26,74 22,56 26,36 Z" fill="%23292524"/>
      <!-- Face -->
      <path d="M35,40 C35,60 40,68 50,70 C60,68 65,60 65,40 Z" fill="%23fff1f2"/>
      <!-- Flower Pupil Eyes -->
      <ellipse cx="42" cy="52" rx="4" ry="4.5" fill="%23ea580c"/>
      <polygon points="42,49 43.5,52 45,52 43.8,53.5 44.5,55 42,54 39.5,55 40.2,53.5 39,52 40.5,52" fill="%23fef08a"/>
      <ellipse cx="58" cy="52" rx="4" ry="4.5" fill="%23ea580c"/>
      <polygon points="58,49 59.5,52 61,52 59.8,53.5 60.5,55 58,54 55.5,55 56.2,53.5 55,52 56.5,52" fill="%23fef08a"/>
      <!-- Mischievous Cat Smile -->
      <path d="M45,62 Q48,65 50,62 Q52,65 55,62" fill="none" stroke="%23c2410c" stroke-width="2" stroke-linecap="round"/>
      <!-- Cute Ghost Buddy on Bottom Right -->
      <ellipse cx="78" cy="74" rx="10" ry="7" fill="%23ffffff" opacity="0.85"/>
      <circle cx="75" cy="73" r="1.5" fill="%231e293b"/>
      <circle cx="81" cy="73" r="1.5" fill="%231e293b"/>
    </svg>`,
  },
  {
    id: 'venti',
    name: '温迪',
    title: '风之诗人 · 巴巴托斯',
    element: 'Anemo',
    elementName: '风元素',
    elementColor: '#4fcda6',
    quote: '若你困于无风之地，我便为你奏响高天之歌。',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="anemoBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%23f0fdfa"/>
          <stop offset="100%" stop-color="%23ccfbf1"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23anemoBg)" stroke="%232dd4bf" stroke-width="3"/>
      <!-- Green Beret with White Feather -->
      <ellipse cx="44" cy="22" rx="22" ry="10" transform="rotate(-15 44 22)" fill="%230f766e"/>
      <polygon points="26,14 18,6 28,12" fill="%23ffffff"/>
      <!-- Twin Braids with Glowing Cyan Tips -->
      <path d="M28,38 C30,22 70,22 72,38 C76,58 72,72 50,76 C28,72 24,58 28,38 Z" fill="%23134e4a"/>
      <path d="M22,54 C20,68 24,78 30,84 C26,76 26,64 22,54 Z" fill="%232dd4bf"/>
      <path d="M78,54 C80,68 76,78 70,84 C74,76 74,64 78,54 Z" fill="%232dd4bf"/>
      <!-- Face -->
      <path d="M35,40 C35,58 40,68 50,70 C60,68 65,58 65,40 Z" fill="%23fffbeb"/>
      <!-- Cyan Turquoise Eyes -->
      <ellipse cx="42" cy="51" rx="3.5" ry="4.5" fill="%2314b8a6"/>
      <circle cx="43" cy="49" r="1.2" fill="%23ffffff"/>
      <ellipse cx="58" cy="51" rx="3.5" ry="4.5" fill="%2314b8a6"/>
      <circle cx="59" cy="49" r="1.2" fill="%23ffffff"/>
      <!-- Carefree Smile -->
      <path d="M46,62 Q50,66 54,62" fill="none" stroke="%230f766e" stroke-width="2" stroke-linecap="round"/>
      <!-- Cecilia Flower -->
      <circle cx="50" cy="85" r="4" fill="%23ffffff" stroke="%235eead4" stroke-width="1.5"/>
    </svg>`,
  },
  {
    id: 'kazuha',
    name: '枫原万叶',
    title: '白鹤归野 · 浪客之道',
    element: 'Anemo',
    elementName: '风元素',
    elementColor: '#d96b52',
    quote: '飘摇风雨中，带刀归来人。',
    avatarSvg: `data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
      <defs>
        <radialGradient id="kazuhaBg" cx="50%" cy="40%" r="60%">
          <stop offset="0%" stop-color="%23fdf2e9"/>
          <stop offset="100%" stop-color="%23fee2e2"/>
        </radialGradient>
      </defs>
      <circle cx="50" cy="50" r="48" fill="url(%23kazuhaBg)" stroke="%23f87171" stroke-width="3"/>
      <!-- Platinum Blonde Hair with Crimson Streak -->
      <path d="M28,36 C30,20 70,20 72,36 C76,56 72,74 50,78 C28,74 24,56 28,36 Z" fill="%23f5f5f4"/>
      <!-- Red Streak on Left Hair -->
      <path d="M34,30 C32,44 33,56 36,66 C35,56 36,44 38,32 Z" fill="%23dc2626"/>
      <!-- Face -->
      <path d="M35,38 C35,58 40,68 50,72 C60,68 65,58 65,38 Z" fill="%23fef2f2"/>
      <!-- Ruby Red Maple Eyes -->
      <ellipse cx="42" cy="50" rx="3.5" ry="4.5" fill="%23b91c1c"/>
      <circle cx="43" cy="48" r="1.2" fill="%23ffffff"/>
      <ellipse cx="58" cy="50" rx="3.5" ry="4.5" fill="%23b91c1c"/>
      <circle cx="59" cy="48" r="1.2" fill="%23ffffff"/>
      <!-- Serene Samurai Mouth -->
      <line x1="47" y1="62" x2="53" y2="62" stroke="%23991b1b" stroke-width="1.5" stroke-linecap="round"/>
      <!-- Maple Leaf Motif -->
      <path d="M50,78 L53,82 L58,80 L55,85 L58,88 L52,87 L50,92 L48,87 L42,88 L45,85 L42,80 L47,82 Z" fill="%23ef4444"/>
    </svg>`,
  },
];
