/* eslint-disable no-useless-escape */
const quoiRegexs = [
    // Quoi
    /qu+o+i+[\ ?]*\?*$/,
    /qu+o+i+[\ +]*\?*$/,

    // Koa
    /ko+a+[\ ?]*\?*$/,
    /ko+a+[\ +]*\?*$/,

    // Qoa
    /q+o+a+[\ ?]*\?*$/,
    /q+o+a+[\ +]*\?*$/,

    // Koi
    /ko+i+[\ ?]*\?*$/,
    /ko+i+[\ +]*\?*$/,

    // Kwa
    /kw+a+[\ ?]*\?*$/,
    /kw+a+[\ +]*\?*$/,

    // Kewa
    /ke+w+a+[\ ?]*\?*$/,
    /ke+w+a+[\ +]*\?*$/,

    // Koua
    /ko+u+a+[\ ?]*\?*$/,
    /ko+u+a+[\ +]*\?*$/,
    /kou+a+[\ ?]*\?*$/,
    /kou+a+[\ +]*\?*$/,

    // Kewa
    /ke+u+a+[\ ?]*\?*$/,
    /ke+u+a+[\ +]*\?*$/,
    /keu+a+[\ ?]*\?*$/,
    /keu+a+[\ +]*\?*$/,

    // Koua
    /ko+u+a+[\ ?]*\?*$/,
    /ko+u+a+[\ +]*\?*$/,
    /kou+a+[\ ?]*\?*$/,
    /kou+a+[\ +]*\?*$/,
]

export const possibleResponses = [
    {
        response: "Feur.",
        probability: 70
    },
    {
        response: "coubeh.",
        probability: 10
    },
    {
        response: "la",
        probability: 10
    },
    {
        response: "drilatère.",
        probability: 10
    },
]


export function detectFeur(message: string): boolean {
    for (const regex of quoiRegexs) {
        if (regex.test(message.toLowerCase())) {
            return true
        }
    }
    return false
}

export function generateResponse(): string {
    const random = Math.random() * 100
    let cumulativeProbability = 0
    for (const response of possibleResponses) {
        cumulativeProbability += response.probability
        if (random <= cumulativeProbability) {
            return response.response
        }
    }
    return "Feur."
}