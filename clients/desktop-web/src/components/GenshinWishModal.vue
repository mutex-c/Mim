<script setup lang="ts">
import { ref, computed } from 'vue';
import { Sparkles, X, User, ShieldAlert, MessageSquare } from 'lucide-vue-next';
import { playGenshinClick, playPrimogemSparkle, playWishChord, playDropReveal } from '../utils/genshinAudio';
import { GENSHIN_CHARACTERS } from '../assets/genshin/characters';

import splashZhongli from '../assets/genshin/images/splash_zhongli.png';
import splashFurina from '../assets/genshin/images/splash_furina.png';
import splashRaiden from '../assets/genshin/images/splash_raiden.png';
import splashHutao from '../assets/genshin/images/splash_hutao.png';
import splashVodyanitsa from '../assets/vodyanitsa/images/splash_vodyanitsa.png';
import iconConch from '../assets/vodyanitsa/dock/nav_calls_conch.png';
import iconLily from '../assets/vodyanitsa/dock/nav_contacts_lily.png';
import iconGrimoire from '../assets/vodyanitsa/dock/nav_grimoire.png';
import iconWing from '../assets/vodyanitsa/dock/nav_chats_wave_wing.png';
import iconGear from '../assets/vodyanitsa/dock/nav_settings_gear.png';
import iconScreen from '../assets/vodyanitsa/dock/nav_screen.png';
import iconBtnHorn from '../assets/vodyanitsa/images/icon_btn_horn.png';
import iconBtnSticker from '../assets/vodyanitsa/images/icon_btn_sticker.png';
import { useThemeStore } from '../stores/theme';

const props = defineProps<{
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'share', payload: { bannerName: string; pulls: WishDrop[]; summaryText: string }): void;
}>();

export interface WishDrop {
  id: string;
  name: string;
  rarity: 3 | 4 | 5;
  type: 'character' | 'weapon';
  element?: string;
  elementColor?: string;
  avatarSvg?: string;
  weaponType?: string;
}

// Banners
interface Banner {
  id: string;
  name: string;
  featuredName: string;
  featuredTitle: string;
  element: string;
  elementColor: string;
  color: string;
  characterId: string;
  splashImg: string;
  description: string;
}

const themeStore = useThemeStore();

const genshinBanners: Banner[] = [
  {
    id: 'zhongli',
    name: '陵薮市朝',
    featuredName: '钟离',
    featuredTitle: '尘世闲游 · 岩王帝君',
    element: '岩',
    elementColor: '#eab308',
    color: '#d97706',
    characterId: 'zhongli',
    splashImg: splashZhongli,
    description: '限定五星角色【尘世闲游·钟离(岩)】祈愿抽取概率大幅提升！',
  },
  {
    id: 'furina',
    name: '众水向往',
    featuredName: '芙宁娜',
    featuredTitle: '不休独舞 · 水神',
    element: '水',
    elementColor: '#38bdf8',
    color: '#0284c7',
    characterId: 'furina',
    splashImg: splashFurina,
    description: '限定五星角色【不休独舞·芙宁娜(水)】祈愿抽取概率大幅提升！',
  },
  {
    id: 'raiden',
    name: '一心净土',
    featuredName: '雷电将军',
    featuredTitle: '一心净土 · 雷神',
    element: '雷',
    elementColor: '#a855f7',
    color: '#7c3aed',
    characterId: 'raiden',
    splashImg: splashRaiden,
    description: '限定五星角色【一心净土·雷电将军(雷)】祈愿抽取概率大幅提升！',
  },
  {
    id: 'hutao',
    name: '赤团开时',
    featuredName: '胡桃',
    featuredTitle: '雪霁梅香 · 往生堂堂主',
    element: '火',
    elementColor: '#ef4444',
    color: '#b91c1c',
    characterId: 'hutao',
    splashImg: splashHutao,
    description: '限定五星角色【雪霁梅香·胡桃(火)】祈愿抽取概率大幅提升！',
  },
];

const vodyanitsaBanners: Banner[] = [
  {
    id: 'vodyanitsa',
    name: '幽歌萦渊',
    featuredName: '沃雅妮莎',
    featuredTitle: '幽歌萦渊 · 首席女高音',
    element: '水',
    elementColor: '#38bdf8',
    color: '#0284c7',
    characterId: 'vodyanitsa',
    splashImg: splashVodyanitsa,
    description: '科洛列夫茨基剧团首席女高音【幽歌萦渊·沃雅妮莎(水)】特约祈愿！限定抽取概率大幅提升！',
  },
];

const activeBannersList = computed(() => {
  return themeStore.currentPalette === 'vodyanitsa' ? vodyanitsaBanners : genshinBanners;
});

const currentBannerIdx = ref(0);
const activeBanner = computed(() => {
  const list = activeBannersList.value;
  return list[currentBannerIdx.value] || list[0];
});

// Primogems & Fates
const primogems = ref(16000);
const intertwinedFates = ref(10);

const addPrimogems = () => {
  primogems.value += 1600;
  playPrimogemSparkle();
};

// Wish States
const isWishing = ref(false);
const wishPhase = ref<'meteor' | 'results' | 'none'>('none');
const currentWishResults = ref<WishDrop[]>([]);
const hasFiveStarInWish = ref(false);

// Drop pools for Genshin Mode
const poolFourStars: WishDrop[] = [
  { id: 'bennett', name: '班尼特', rarity: 4, type: 'character', element: '火', elementColor: '#ef4444' },
  { id: 'xingqiu', name: '行秋', rarity: 4, type: 'character', element: '水', elementColor: '#38bdf8' },
  { id: 'xiangling', name: '香菱', rarity: 4, type: 'character', element: '火', elementColor: '#ef4444' },
  { id: 'fischl', name: '菲谢尔', rarity: 4, type: 'character', element: '雷', elementColor: '#a855f7' },
  { id: 'sucrose', name: '砂糖', rarity: 4, type: 'character', element: '风', elementColor: '#2dd4bf' },
  { id: 'w_rust', name: '弓·弓藏', rarity: 4, type: 'weapon', weaponType: '弓' },
  { id: 'w_widsith', name: '法器·流浪乐章', rarity: 4, type: 'weapon', weaponType: '法器' },
  { id: 'w_favonius', name: '单手剑·西风剑', rarity: 4, type: 'weapon', weaponType: '单手剑' },
];

const poolThreeStars: WishDrop[] = [
  { id: 'w_harbinger', name: '黎明神剑', rarity: 3, type: 'weapon', weaponType: '单手剑' },
  { id: 'w_ttds', name: '讨龙继承者的英杰谭', rarity: 3, type: 'weapon', weaponType: '法器' },
  { id: 'w_slingshot', name: '弹弓', rarity: 3, type: 'weapon', weaponType: '弓' },
  { id: 'w_tassel', name: '黑缨枪', rarity: 3, type: 'weapon', weaponType: '长柄武器' },
  { id: 'w_debate', name: '以理服人', rarity: 3, type: 'weapon', weaponType: '双手剑' },
];

// Drop pools for 100% Pure Vodyanitsa Mode
const vodyanitsaFourStars: WishDrop[] = [
  { id: 'v_aria', name: '水妖咏叹调', rarity: 4, type: 'weapon', weaponType: '法器', avatarSvg: iconBtnHorn },
  { id: 'v_bow', name: '科洛列夫茨基之音', rarity: 4, type: 'weapon', weaponType: '弓', avatarSvg: iconWing },
  { id: 'v_score', name: '冰雪双姝赋格', rarity: 4, type: 'weapon', weaponType: '乐谱', avatarSvg: iconGrimoire },
  { id: 'v_horn', name: '唤春金螺', rarity: 4, type: 'weapon', weaponType: '道具', avatarSvg: iconConch },
  { id: 'v_feather', name: '纯白羽翼之冠', rarity: 4, type: 'weapon', weaponType: '圣遗物', avatarSvg: iconLily },
  { id: 'v_gramophone', name: '剧院留声机', rarity: 4, type: 'weapon', weaponType: '道具', avatarSvg: iconScreen },
];

const vodyanitsaThreeStars: WishDrop[] = [
  { id: 'v_practice', name: '剧团练习总谱', rarity: 3, type: 'weapon', weaponType: '乐谱', avatarSvg: iconGrimoire },
  { id: 'v_spotlight', name: '聚光透镜', rarity: 3, type: 'weapon', weaponType: '道具', avatarSvg: iconGear },
  { id: 'v_tear', name: '水妖之泪', rarity: 3, type: 'weapon', weaponType: '晶石', avatarSvg: iconBtnSticker },
  { id: 'v_reed', name: '双簧管芦哨', rarity: 3, type: 'weapon', weaponType: '乐器', avatarSvg: iconConch },
  { id: 'v_petal', name: '冰湖水仙花瓣', rarity: 3, type: 'weapon', weaponType: '材料', avatarSvg: iconLily },
];

function performWish(count: 1 | 10) {
  const cost = count * 160;
  if (primogems.value < cost && intertwinedFates.value < count) {
    primogems.value += 3200; // Free refill on shortage
  }
  if (intertwinedFates.value >= count) {
    intertwinedFates.value -= count;
  } else {
    primogems.value = Math.max(0, primogems.value - cost);
  }

  isWishing.value = true;
  wishPhase.value = 'meteor';

  const results: WishDrop[] = [];
  let got5Star = false;
  const isVody = themeStore.currentPalette === 'vodyanitsa';

  for (let i = 0; i < count; i++) {
    const roll = Math.random();
    // 25% chance 5-star for exciting interactive demo! Guaranteed 4-star in 10-pull
    if (roll < 0.25 || (count === 10 && i === 9 && !got5Star && !results.some(r => r.rarity >= 4))) {
      // 5-Star Drop!
      got5Star = true;
      if (isVody) {
        results.push({
          id: 'vodyanitsa',
          name: '沃雅妮莎',
          rarity: 5,
          type: 'character',
          element: '水',
          elementColor: '#38bdf8',
          avatarSvg: splashVodyanitsa,
        });
      } else {
        const char = GENSHIN_CHARACTERS.find(c => c.id === activeBanner.value.characterId) || GENSHIN_CHARACTERS[1];
        results.push({
          id: char.id,
          name: char.name,
          rarity: 5,
          type: 'character',
          element: char.elementName.replace('之眼', ''),
          elementColor: char.elementColor,
          avatarSvg: activeBanner.value.splashImg,
        });
      }
    } else if (roll < 0.65 || (count === 10 && i === 9 && !got5Star)) {
      // 4-Star Drop
      if (isVody) {
        const picked = vodyanitsaFourStars[Math.floor(Math.random() * vodyanitsaFourStars.length)];
        results.push({ ...picked });
      } else {
        const picked = poolFourStars[Math.floor(Math.random() * poolFourStars.length)];
        results.push({ ...picked });
      }
    } else {
      // 3-Star Drop
      if (isVody) {
        const picked = vodyanitsaThreeStars[Math.floor(Math.random() * vodyanitsaThreeStars.length)];
        results.push({ ...picked });
      } else {
        const picked = poolThreeStars[Math.floor(Math.random() * poolThreeStars.length)];
        results.push({ ...picked });
      }
    }
  }

  hasFiveStarInWish.value = got5Star;
  playWishChord(got5Star);

  // Meteor animation ends in 1.4s, then show results
  setTimeout(() => {
    wishPhase.value = 'results';
    currentWishResults.value = results;
  }, 1400);
}

function handleClose() {
  playGenshinClick();
  wishPhase.value = 'none';
  isWishing.value = false;
  emit('close');
}

function handleShareToChat() {
  playPrimogemSparkle();
  const drops = currentWishResults.value;
  const fiveStars = drops.filter(d => d.rarity === 5).map(d => d.name);
  const fourStars = drops.filter(d => d.rarity === 4).map(d => d.name);
  const isVody = themeStore.currentPalette === 'vodyanitsa';

  let summary = isVody
    ? `【唤春角笛·剧院特约祈愿报告】`
    : `【${activeBanner.value.name}·祈愿报告】`;

  if (fiveStars.length > 0) {
    summary += isVody
      ? ` ✦ 歌声共鸣！剧院首席女高音【${fiveStars.join('、')}】华丽登台！`
      : ` ✦ 金色传说！斩获五星：${fiveStars.join('、')}！`;
  }
  if (fourStars.length > 0) {
    summary += isVody
      ? ` ★ 剧院珍藏：${fourStars.join('、')}`
      : ` ★ 四星伙伴：${fourStars.join('、')}`;
  }
  if (fiveStars.length === 0 && fourStars.length === 0) {
    summary += isVody
      ? ` 获得剧院练习道具补给 x${drops.length}`
      : ` 获得 3星武器补给 x${drops.length}`;
  }

  emit('share', {
    bannerName: isVody ? '幽歌萦渊' : activeBanner.value.name,
    pulls: drops,
    summaryText: summary,
  });

  handleClose();
}
</script>

<template>
  <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md select-none font-sans">
    <!-- Main Wish Window Frame -->
    <div
      class="relative w-[95vw] max-w-5xl h-[88vh] max-h-[760px] rounded-3xl border-2 flex flex-col overflow-hidden transition-all duration-300"
      :class="themeStore.currentPalette === 'vodyanitsa'
        ? 'bg-gradient-to-b from-[#0a1835] via-[#071127] to-[#040917] border-cyan-400/80 shadow-[0_0_60px_rgba(56,189,248,0.35)]'
        : 'bg-gradient-to-b from-[#1c1d29] via-[#161722] to-[#0e0f17] border-[#caa159]/70 shadow-[0_0_50px_rgba(202,161,89,0.35)]'"
    >
      
      <!-- Top Filigree Header Bar -->
      <div
        class="relative h-14 border-b flex items-center justify-between px-6 z-20 transition-colors"
        :class="themeStore.currentPalette === 'vodyanitsa'
          ? 'bg-gradient-to-r from-[#0d1d3d] via-[#122852] to-[#0d1d3d] border-cyan-400/40'
          : 'bg-gradient-to-r from-[#202133] via-[#2a2c42] to-[#202133] border-[#caa159]/40'"
      >
        <!-- Left Ornaments & Title -->
        <div class="flex items-center gap-3">
          <div
            class="w-8 h-8 rounded-full border flex items-center justify-center shadow-inner overflow-hidden"
            :class="themeStore.currentPalette === 'vodyanitsa'
              ? 'border-cyan-400/60 bg-[#09152e]'
              : 'border-[#caa159] bg-[#1a1b26]'"
          >
            <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconBtnHorn" alt="" class="w-5 h-5 object-contain" />
            <span v-else class="text-amber-400 text-lg">✦</span>
          </div>
          <span
            class="text-lg font-bold tracking-widest drop-shadow-[0_1px_3px_rgba(0,0,0,0.8)]"
            :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-100' : 'text-[#f3e7c4]'"
          >
            {{ themeStore.currentPalette === 'vodyanitsa' ? '科洛列夫茨基剧团 · 唤春角笛特约祈愿' : '提瓦特祈愿 · 角色活动祈愿' }}
          </span>
        </div>

        <!-- Banner Selector Tabs (Only show for Genshin mode with multiple banners) -->
        <div
          v-if="activeBannersList.length > 1"
          class="hidden md:flex items-center gap-2 bg-[#12131c]/80 p-1 rounded-xl border border-[#caa159]/30"
        >
          <button
            v-for="(b, idx) in activeBannersList"
            :key="b.id"
            @click="currentBannerIdx = idx; playGenshinClick()"
            class="px-3.5 py-1 text-xs font-semibold rounded-lg transition-all duration-200 cursor-pointer"
            :class="currentBannerIdx === idx ? 'bg-gradient-to-r from-[#d97706] to-[#b45309] text-white shadow-md' : 'text-[#f3e7c4]/70 hover:text-white'"
          >
            {{ b.name }}
          </button>
        </div>
        <div v-else-if="themeStore.currentPalette === 'vodyanitsa'" class="hidden md:flex items-center gap-1.5 px-3 py-1 rounded-xl bg-cyan-950/70 border border-cyan-400/40 text-cyan-200 text-xs font-bold tracking-wider">
          <Sparkles class="w-3.5 h-3.5 text-cyan-400 inline-icon" />
          <span>幽歌萦渊 · 沃雅妮莎特邀登台</span>
        </div>

        <!-- Primogem Counter & Close Button -->
        <div class="flex items-center gap-4">
          <!-- Fate -->
          <div class="flex items-center gap-1.5 px-3 py-1 bg-black/50 border border-[#caa159]/50 rounded-full text-xs font-medium text-[#f3e7c4]">
            <Sparkles class="w-3.5 h-3.5 text-pink-400 inline-icon" />
            <span>{{ intertwinedFates }}</span>
          </div>

          <!-- Primogem with + Button -->
          <div class="flex items-center gap-1.5 pl-3 pr-1 py-1 bg-black/50 border border-[#caa159]/50 rounded-full text-xs font-medium text-[#f3e7c4]">
            <span class="text-sky-400">✦</span>
            <span class="font-bold text-amber-200">{{ primogems.toLocaleString() }}</span>
            <button
              @click="addPrimogems"
              title="获取原石补给 (+1600)"
              class="ml-1 w-5 h-5 rounded-full bg-amber-500/30 hover:bg-amber-500/60 text-amber-200 flex items-center justify-center text-xs font-bold transition-all"
            >
              +
            </button>
          </div>

          <!-- Close -->
          <button
            @click="handleClose"
            class="w-8 h-8 rounded-full border border-[#caa159]/60 bg-[#2d1b22] text-[#fca5a5] hover:bg-red-900/60 flex items-center justify-center font-bold text-sm transition-all"
          >
            <X class="w-4 h-4 inline-icon" />
          </button>
        </div>
      </div>

      <!-- Banner Body Display Area -->
      <div v-if="wishPhase === 'none'" class="relative flex-1 flex flex-col justify-between p-8 overflow-hidden">
        <!-- Background Banner Art / Ambient Gradient -->
        <div
          class="absolute inset-0 opacity-20 pointer-events-none transition-all duration-700"
          :style="{
            background: `radial-gradient(circle at 65% 40%, ${activeBanner.color}, transparent 65%)`
          }"
        />

        <!-- Corner Filigree Decors -->
        <div class="absolute top-4 left-4 w-12 h-12 border-t-2 border-l-2 pointer-events-none" :class="themeStore.currentPalette === 'vodyanitsa' ? 'border-cyan-400/60' : 'border-[#caa159]/60'" />
        <div class="absolute top-4 right-4 w-12 h-12 border-t-2 border-r-2 pointer-events-none" :class="themeStore.currentPalette === 'vodyanitsa' ? 'border-cyan-400/60' : 'border-[#caa159]/60'" />
        <div class="absolute bottom-4 left-4 w-12 h-12 border-b-2 border-l-2 pointer-events-none" :class="themeStore.currentPalette === 'vodyanitsa' ? 'border-cyan-400/60' : 'border-[#caa159]/60'" />
        <div class="absolute bottom-4 right-4 w-12 h-12 border-b-2 border-r-2 pointer-events-none" :class="themeStore.currentPalette === 'vodyanitsa' ? 'border-cyan-400/60' : 'border-[#caa159]/60'" />

        <!-- Banner Info Block -->
        <div class="relative z-10 max-w-lg mt-4 space-y-3">
          <div
            class="inline-flex items-center gap-2 px-3 py-1 rounded-md text-xs font-semibold tracking-wider"
            :class="themeStore.currentPalette === 'vodyanitsa'
              ? 'bg-[#0a2046]/90 border border-cyan-400/40 text-cyan-200'
              : 'bg-[#25273b]/90 border border-[#caa159]/40 text-amber-300'"
          >
            <span>{{ activeBanner.element }}元素限定</span>
            <span>★★★★★</span>
          </div>

          <h1 class="text-4xl md:text-5xl font-black text-[#fdf8e2] drop-shadow-[0_2px_8px_rgba(0,0,0,0.9)] tracking-wide">
            {{ activeBanner.featuredName }}
          </h1>

          <p class="text-sm font-semibold tracking-widest" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-300' : 'text-amber-300/90'">
            {{ activeBanner.featuredTitle }}
          </p>

          <p class="text-xs text-[#d1d5db]/80 leading-relaxed max-w-md pt-2">
            {{ activeBanner.description }}
          </p>
        </div>

        <!-- Featured Character Showcase Graphic -->
        <div class="absolute right-8 md:right-16 top-6 bottom-20 w-80 md:w-[420px] flex items-center justify-center pointer-events-none z-0">
          <div class="relative w-72 h-[380px] md:w-96 md:h-[480px] flex items-center justify-center p-2">
            <img
              :src="activeBanner.splashImg"
              class="w-full h-full object-contain drop-shadow-[0_0_40px_rgba(56,189,248,0.5)] transform hover:scale-102 transition-transform duration-500 rounded-3xl"
              :alt="activeBanner.featuredName"
            />
          </div>
        </div>

        <!-- Bottom Action Pull Buttons -->
        <div class="relative z-10 flex items-center justify-between pt-6 border-t" :class="themeStore.currentPalette === 'vodyanitsa' ? 'border-cyan-400/30' : 'border-[#caa159]/30'">
          <!-- Left Pity & Details Note -->
          <div class="flex items-center gap-2 text-xs text-[#e5e7eb]/70">
            <span class="w-2 h-2 rounded-full animate-ping" :class="themeStore.currentPalette === 'vodyanitsa' ? 'bg-cyan-400' : 'bg-amber-400'" />
            <span>十连祈愿必定获得至少1个 4★ 或以上专属道具/角色</span>
          </div>

          <!-- Pull Buttons -->
          <div class="flex items-center gap-4">
            <!-- 1 Pull -->
            <button
              @click="performWish(1)"
              class="relative px-6 py-2.5 rounded-xl border shadow-lg transition-all flex flex-col items-center group active:scale-95 cursor-pointer"
              :class="themeStore.currentPalette === 'vodyanitsa'
                ? 'border-cyan-400/60 bg-gradient-to-b from-[#0e2246] to-[#071329] text-cyan-100 hover:border-cyan-300'
                : 'border-[#caa159]/70 bg-gradient-to-b from-[#2e3146] to-[#1a1b28] hover:border-amber-400 text-[#f3e7c4]'"
            >
              <span class="text-xs font-bold tracking-wider">祈愿 1 次</span>
              <span class="text-[10px] flex items-center gap-1 font-mono" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-300' : 'text-amber-300'">
                ✦ 160
              </span>
            </button>

            <!-- 10 Pull -->
            <button
              @click="performWish(10)"
              class="relative px-8 py-2.5 rounded-xl border-2 transition-all flex flex-col items-center group active:scale-95 font-bold cursor-pointer"
              :class="themeStore.currentPalette === 'vodyanitsa'
                ? 'border-cyan-300 bg-gradient-to-b from-cyan-600 via-sky-600 to-blue-800 text-white shadow-[0_0_25px_rgba(56,189,248,0.5)] hover:from-cyan-500 hover:to-blue-700'
                : 'border-[#eab308] bg-gradient-to-b from-[#d97706] to-[#92400e] text-white shadow-[0_0_25px_rgba(234,179,8,0.5)] hover:from-[#f59e0b] hover:to-[#b45309]'"
            >
              <span class="text-sm font-black tracking-widest flex items-center gap-1.5">
                <span :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-200' : 'text-amber-200'">★</span> 祈愿 10 次
              </span>
              <span class="text-[11px] flex items-center gap-1 font-mono" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-100' : 'text-amber-100'">
                ✦ 1600
              </span>
            </button>
          </div>
        </div>
      </div>

      <!-- Meteor Animation Overlay -->
      <div v-else-if="wishPhase === 'meteor'" class="absolute inset-0 z-30 flex flex-col items-center justify-center bg-black/95 overflow-hidden">
        <!-- Shooting Star Ray -->
        <div
          class="relative w-full h-2 rounded-full overflow-visible flex items-center justify-center animate-pulse"
        >
          <!-- Golden meteor trail if 5 star, purple/cyan if 4 star -->
          <div
            class="h-3 md:h-4 w-3/4 rounded-full filter blur-sm transition-all duration-300"
            :class="hasFiveStarInWish
              ? (themeStore.currentPalette === 'vodyanitsa' ? 'bg-gradient-to-r from-transparent via-[#7dd3fc] to-[#38bdf8] shadow-[0_0_50px_#38bdf8]' : 'bg-gradient-to-r from-transparent via-[#fef08a] to-[#eab308] shadow-[0_0_50px_#fde047]')
              : 'bg-gradient-to-r from-transparent via-[#e9d5ff] to-[#a855f7] shadow-[0_0_40px_#c084fc]'"
          />
          <div
            class="absolute w-8 h-8 rounded-full animate-spin"
            :class="hasFiveStarInWish
              ? (themeStore.currentPalette === 'vodyanitsa' ? 'bg-[#f0f9ff] shadow-[0_0_60px_#38bdf8]' : 'bg-[#fffbeb] shadow-[0_0_60px_#facc15]')
              : 'bg-[#faf5ff] shadow-[0_0_50px_#9333ea]'"
          />
        </div>

        <div class="mt-8 text-center space-y-2">
          <p class="text-xl md:text-2xl font-black tracking-widest" :class="hasFiveStarInWish ? (themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-300 animate-bounce' : 'text-amber-300 animate-bounce') : 'text-purple-300'">
            {{ hasFiveStarInWish ? (themeStore.currentPalette === 'vodyanitsa' ? '✦✦✦ 水妖咏叹 · 沃雅妮莎降临 ✦✦✦' : '✦✦✦ 金色传说 ✦✦✦') : '★ 剧院珍藏 · 星辉闪耀 ★' }}
          </p>
          <p class="text-xs text-[#9ca3af] font-mono">
            {{ themeStore.currentPalette === 'vodyanitsa' ? '水妖的歌声正在激荡，科洛列夫茨基大剧院的帷幕为你升起……' : '命运的丝线正在交织，提瓦特的星空为你指引……' }}
          </p>
        </div>
      </div>

      <!-- Results Display View -->
      <div
        v-else-if="wishPhase === 'results'"
        class="relative flex-1 flex flex-col justify-between p-6 overflow-hidden"
        :class="themeStore.currentPalette === 'vodyanitsa' ? 'bg-gradient-to-b from-[#091736] to-[#040a18]' : 'bg-gradient-to-b from-[#181926] to-[#0c0d14]'"
      >
        <!-- Header -->
        <div class="text-center pt-2">
          <h2 class="text-xl font-bold tracking-widest drop-shadow" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-100' : 'text-[#f5eed6]'">
            {{ themeStore.currentPalette === 'vodyanitsa' ? '祈愿结果 · 科洛列夫茨基剧团的馈赠' : '祈愿结果 · 提瓦特的馈赠' }}
          </h2>
          <p class="text-xs font-mono mt-0.5" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-300' : 'text-amber-400/80'">
            {{ hasFiveStarInWish ? (themeStore.currentPalette === 'vodyanitsa' ? '✦ 恭喜！剧团首席女高音 · 沃雅妮莎华丽登台！✦' : '✦ 恭喜旅行者斩获五星伙伴！✦') : '祈愿完成' }}
          </p>
        </div>

        <!-- Drops Grid -->
        <div class="flex-1 flex items-center justify-center p-4">
          <div class="grid grid-cols-5 md:grid-cols-10 gap-2.5 max-w-4xl w-full">
            <div
              v-for="(item, idx) in currentWishResults"
              :key="idx"
              class="relative rounded-xl border flex flex-col items-center justify-between p-2 h-44 md:h-52 overflow-hidden transition-all transform hover:-translate-y-2 duration-200"
              :class="{
                'bg-gradient-to-b from-[#78350f]/80 via-[#451a03]/90 to-[#1e1b4b] border-[#facc15] shadow-[0_0_20px_rgba(250,204,21,0.45)] ring-2 ring-amber-400/50': item.rarity === 5,
                'bg-gradient-to-b from-[#4c1d95]/70 via-[#2e1065]/80 to-[#111827] border-[#c084fc] shadow-[0_0_15px_rgba(168,85,247,0.3)]': item.rarity === 4,
                'bg-gradient-to-b from-[#1e293b]/70 to-[#0f172a] border-[#38bdf8]/40': item.rarity === 3,
              }"
            >
              <!-- Element / Type Badge -->
              <div class="w-full flex items-center justify-between text-[10px] px-1 font-semibold">
                <span v-if="item.element" :style="{ color: item.elementColor }">{{ item.element }}</span>
                <span v-else class="text-slate-400">{{ item.weaponType }}</span>
                <span :class="item.rarity === 5 ? 'text-amber-300' : item.rarity === 4 ? 'text-purple-300' : 'text-sky-300'">
                  {{ '★'.repeat(item.rarity) }}
                </span>
              </div>

              <!-- Artwork / Icon -->
              <div class="flex-1 flex items-center justify-center my-1 w-full">
                <img
                  v-if="item.avatarSvg"
                  :src="item.avatarSvg"
                  class="w-16 h-16 md:w-20 md:h-20 object-contain drop-shadow"
                  :alt="item.name"
                />
                <div v-else class="w-14 h-14 rounded-full flex items-center justify-center bg-black/40 border border-white/20">
                  <User v-if="item.type === 'character'" class="w-6 h-6 text-amber-200 inline-icon" />
                  <ShieldAlert v-else class="w-6 h-6 text-amber-200 inline-icon" />
                </div>
              </div>

              <!-- Name -->
              <div class="w-full text-center">
                <p class="text-xs font-bold truncate text-[#fef9c3]">
                  {{ item.name }}
                </p>
                <span class="text-[9px] text-[#9ca3af]">
                  {{ item.type === 'character' ? '角色' : '武器' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Action Footer -->
        <div class="flex items-center justify-between pt-4 border-t border-[#caa159]/30">
          <button
            @click="wishPhase = 'none'"
            class="px-5 py-2 rounded-xl border border-[#caa159]/50 text-xs font-semibold text-[#f3e7c4] hover:bg-white/10 transition-all cursor-pointer"
          >
            ← 返回祈愿
          </button>

          <div class="flex items-center gap-3">
            <button
              @click="handleShareToChat"
              class="px-6 py-2 rounded-xl bg-gradient-to-r from-[#22c55e] to-[#15803d] text-white text-xs font-bold shadow-lg hover:from-[#16a34a] hover:to-[#166534] transition-all flex items-center gap-2 cursor-pointer"
            >
              <MessageSquare class="w-3.5 h-3.5 inline-icon" />
              <span>分享祈愿到聊天</span>
            </button>

            <button
              @click="handleClose"
              class="px-6 py-2 rounded-xl bg-gradient-to-r from-[#d97706] to-[#b45309] text-white text-xs font-bold shadow-lg hover:from-[#f59e0b] hover:to-[#d97706] transition-all"
            >
              确认
            </button>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
/* Genshin style custom scrollbar and glow */
</style>
