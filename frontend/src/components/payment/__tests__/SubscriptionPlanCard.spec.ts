import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import { createI18n } from "vue-i18n";
import SubscriptionPlanCard from "../SubscriptionPlanCard.vue";

const msg = (value: string) => () => value;
const named = (ctx: { named: (key: string) => unknown }, key: string) => String(ctx.named(key));

const i18n = createI18n({
  legacy: false,
  locale: "en",
  fallbackWarn: false,
  missingWarn: false,
  messages: {
    en: {
      payment: {
        days: msg("days"),
        models: msg("Models"),
        perMonth: msg("month"),
        perYear: msg("year"),
        planCard: {
          currentPrice: msg("Current price"),
          dailyLimit: msg("Daily"),
          dealHint: (ctx: { named: (key: string) => unknown }) => `Originally ${named(ctx, "original")}, now ${named(ctx, "price")} to start.`,
          discountPercent: (ctx: { named: (key: string) => unknown }) => `${named(ctx, "percent")}% off`,
          durationDays: (ctx: { named: (key: string) => unknown }) => `${named(ctx, "count")} days`,
          durationMonths: (ctx: { named: (key: string) => unknown }) => `${named(ctx, "count")} months`,
          durationWeeks: (ctx: { named: (key: string) => unknown }) => `${named(ctx, "count")} weeks`,
          durationYears: (ctx: { named: (key: string) => unknown }) => `${named(ctx, "count")} years`,
          monthlyLimit: msg("Monthly"),
          originalPrice: msg("Original"),
          quota: msg("Quota"),
          rate: msg("Rate"),
          restorePrice: msg("Restores to"),
          saveAmount: (ctx: { named: (key: string) => unknown }) => `Save ${named(ctx, "amount")}`,
          unlimited: msg("Unlimited"),
          validFor: (ctx: { named: (key: string) => unknown }) => `Valid for ${named(ctx, "duration")}`,
          weeklyLimit: msg("Weekly"),
          limitedOffer: msg("Limited-time offer"),
          limitedOfferUntil: (ctx: { named: (key: string) => unknown }) =>
            `Offer ends ${named(ctx, "time")}; then restores to ${named(ctx, "price")}.`,
          workdayFriendly: msg("Workday-friendly quota"),
          workdayFriendlyDesc: (ctx: { named: (key: string) => unknown }) =>
            `${named(ctx, "daily")} per day and ${named(ctx, "weekly")} per week, tuned for Monday-Friday usage.`,
        },
        subscribeNow: msg("Subscribe now"),
        renewNow: msg("Renew"),
      },
    },
  },
});

const basePlan = {
  id: 1,
  group_id: 10,
  group_platform: "openai",
  name: "Codex Lite",
  description: "Monthly Codex plan",
  price: 99,
  original_price: 149,
  features: ["30-day access"],
  rate_multiplier: 1,
  validity_days: 1,
  validity_unit: "month",
  for_sale: true,
  sort_order: 0,
}

const mountPlanCard = (groupPlatform: string, overrides = {}) =>
  mount(SubscriptionPlanCard, {
    props: {
      plan: {
        ...basePlan,
        group_platform: groupPlatform,
        supported_model_scopes: ["claude", "gemini_text", "gemini_image"],
        ...overrides,
      },
    },
    global: { plugins: [i18n] },
  });

describe("SubscriptionPlanCard", () => {
  it("does not show Antigravity model scopes for OpenAI plans", () => {
    const text = mountPlanCard("openai").text();

    expect(text).not.toContain("Claude");
    expect(text).not.toContain("Gemini");
    expect(text).not.toContain("Imagen");
  });

  it("shows model scopes for Antigravity plans", () => {
    const text = mountPlanCard("antigravity").text();

    expect(text).toContain("Claude");
    expect(text).toContain("Gemini");
    expect(text).toContain("Imagen");
  });

  it("highlights original price, current price, and savings", () => {
    const text = mountPlanCard("openai").text();

    expect(text).toContain("Current price");
    expect(text).toContain("$99");
    expect(text).toContain("Original");
    expect(text).toContain("$149");
    expect(text).toContain("34% off");
    expect(text).toContain("Save $50");
    expect(text).toContain("Originally $149, now $99 to start.");
  });

  it("shows workday-friendly quota messaging for weekday-shaped limits", () => {
    const text = mountPlanCard("openai", {
      daily_limit_usd: 15,
      weekly_limit_usd: 70,
      monthly_limit_usd: 300,
    }).text();

    expect(text).toContain("Daily");
    expect(text).toContain("$15");
    expect(text).toContain("Weekly");
    expect(text).toContain("$70");
    expect(text).toContain("Monthly");
    expect(text).toContain("$300");
    expect(text).toContain("Workday-friendly quota");
    expect(text).toContain("$15 per day and $70 per week");
  });

  it("shows workday-friendly quota messaging for every plan with daily and weekly limits", () => {
    const text = mountPlanCard("openai", {
      name: "Codex Pro",
      daily_limit_usd: 20,
      weekly_limit_usd: 105,
      monthly_limit_usd: 450,
    }).text();

    expect(text).toContain("Workday-friendly quota");
    expect(text).toContain("$20 per day and $105 per week");
  });

  it("highlights active limited-time offers", () => {
    const text = mountPlanCard("openai", {
      price: 99,
      regular_price: 129,
      original_price: 0,
      limited_offer_price: 99,
      limited_offer_expires_at: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
      limited_offer_active: true,
    }).text();

    expect(text).toContain("Limited-time offer");
    expect(text).toContain("Restores to");
    expect(text).toContain("$129");
    expect(text).toContain("Offer ends");
  });

  it("calculates limited-time offer savings against configured original price", () => {
    const text = mountPlanCard("openai", {
      price: 99,
      regular_price: 129,
      original_price: 400,
      limited_offer_price: 99,
      limited_offer_expires_at: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
      limited_offer_active: true,
    }).text();

    expect(text).toContain("75% off");
    expect(text).toContain("Save $301");
    expect(text).toContain("Restores to");
    expect(text).toContain("$129");
    expect(text).toContain("Originally $400, now $99 to start.");
  });
});
