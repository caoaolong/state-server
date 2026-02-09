import dayjs from "dayjs";

/** ISO 时间字符串格式化为本地显示：YYYY-MM-DD HH:mm:ss */
export function formatDateTime(isoOrDate: string | Date | null | undefined): string {
  if (isoOrDate == null || isoOrDate === "") return "";
  const d = dayjs(isoOrDate);
  return d.isValid() ? d.format("YYYY-MM-DD HH:mm:ss") : String(isoOrDate);
}
