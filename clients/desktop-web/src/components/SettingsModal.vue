<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-lg rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col">
      <!-- Modal Header -->
      <div class="p-5 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <Settings class="w-5 h-5 text-blue-600 dark:text-blue-400 inline-icon" />
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">Mim Enterprise System Settings</h3>
        </div>
        <button @click="$emit('close')" class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Settings Content -->
      <div class="p-6 space-y-5 overflow-y-auto max-h-[70vh]">
        <!-- Item 1: Protocol & Transport Architecture -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">Transport Layer Architecture</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono font-medium">
              24-Byte Aligned Binary Frame
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            Network Byte Order (Big-Endian) across all multi-byte fields. Monotonic gap-tolerant SeqId, CRC32 packet integrity validation, CAS concurrency guard for conversation previews.
          </p>
        </div>

        <!-- Item 2: Security & DFA Filter -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">DFA Sensitive Word Filter</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-blue-500/10 text-blue-600 dark:text-blue-400 font-mono font-medium">
              Active (Deterministic Finite Automaton)
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            Real-time keyword filtering for unauthorized leaks, phishing, and abuse. Automatically masks malicious payloads with asterisks prior to database persistence.
          </p>
        </div>

        <!-- Item 3: File Instant Upload (秒传) -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">SHA-256 Storage Deduplication</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-purple-500/10 text-purple-600 dark:text-purple-400 font-mono font-medium">
              Instant Upload (秒传) Active
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            Files are hashed with SHA-256 upon selection. Matching existing file hashes bypass disk writing and return instantaneous zero-latency upload responses.
          </p>
        </div>

        <!-- Item 4: WebRTC Signaling -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">WebRTC Ephemeral Signaling</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-amber-500/10 text-amber-600 dark:text-amber-400 font-mono font-medium">
              Strictly Memory Only
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            Offer, Answer, and ICE Candidate packets (0x0500) route purely in WebSocket memory without polluting message tables. Only finalized call logs are archived.
          </p>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="p-4 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.02] flex justify-end">
        <button
          @click="$emit('close')"
          class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors"
        >
          Done
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Settings, X } from 'lucide-vue-next';

defineEmits(['close']);
</script>
