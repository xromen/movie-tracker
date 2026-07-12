export function convertNumberToCurrency(value: number, currencyType: string, locale: string) {
    if (!value) {
        return ""
    }
    return value.toLocaleString(locale, {
        style: "currency",
        currency: currencyType,
        minimumFractionDigits: 0,
        maximumFractionDigits: 0,
    })
}

export function minsToTimeConverter(mins: number) {
    const days = Math.floor(mins / (60 * 24))
    const hours = Math.floor((mins % (60 * 24)) / 60)
    const minutes = mins % 60

    return {
        days,
        hours,
        minutes,
    }
}