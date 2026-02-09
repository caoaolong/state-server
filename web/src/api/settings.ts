/** 主题 */
export type ThemeOption = "auto" | "light" | "dark";

/** 语言 */
export type LanguageOption = "zh-CN" | "en";

/** 设置表单（设置页） */
export interface SettingsForm {
  appName: string;
  apiBaseUrl: string;
  theme: ThemeOption;
  language: LanguageOption;
  enableNotifications: boolean;
  maxHistoryItems: number;
}

/** 获取设置 - 返回（与 SettingsForm 一致） */
export type SettingsGetResponse = SettingsForm;

/** 保存设置 - 请求（与 SettingsForm 一致） */
export type SettingsSaveRequest = SettingsForm;

/** 保存设置 - 返回 */
export interface SettingsSaveResponse {
  ok: boolean;
}

import { http } from "./request";

const BASE = import.meta.env.VITE_API_BASE ?? "";

/**
 * 获取当前用户/应用设置
 * GET /settings
 */
export async function getSettings(): Promise<SettingsGetResponse> {
  return http<SettingsGetResponse>({
    method: "GET",
    url: `${BASE}/settings`,
  });
}

/**
 * 保存设置
 * PUT /settings
 * Body: SettingsSaveRequest
 */
export async function saveSettings(
  req: SettingsSaveRequest
): Promise<SettingsSaveResponse> {
  return http<SettingsSaveResponse>({
    method: "PUT",
    url: `${BASE}/settings`,
    body: req,
  });
}
