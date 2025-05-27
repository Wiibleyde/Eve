/* eslint-disable no-useless-escape */

// Type definitions for our pattern detection system
type ResponseOption = {
    response: string;
    probability: number;
};

type PatternDetector = {
    name: string;
    regexes: RegExp[];
    responses: ResponseOption[];
    defaultResponse: string;
};

// Quoi detector configuration
const quoiDetector: PatternDetector = {
    name: 'quoi',
    regexes: [
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
        /k+e+w+a+[\ ?]*\?*$/,
        /k+e+w+a+[\ +]*\?*$/,
    ],
    responses: [
        {
            response: 'Feur.',
            probability: 70,
        },
        {
            response: 'coubeh.',
            probability: 10,
        },
        {
            response: 'la 🐨',
            probability: 10,
        },
        {
            response: 'drilatère.',
            probability: 10,
        },
    ],
    defaultResponse: 'Feur.',
};

// Comment detector configuration
const commentDetector: PatternDetector = {
    name: 'comment',
    regexes: [/comment+[\ ?]*\?*$/, /comment+[\ +]*\?*$/, /komen+[\ ?]*\?*$/, /komen+[\ +]*\?*$/],
    responses: [
        {
            response: 'Tateur !',
            probability: 60,
        },
        {
            response: 'dant Cousteau !',
            probability: 10,
        },
        {
            response: 'cement !',
            probability: 20,
        },
    ],
    defaultResponse: 'Tateur !',
};

// Store all detectors in an array for easy access
const allDetectors: PatternDetector[] = [quoiDetector, commentDetector];

// Generic detection function
export function detectPattern(message: string, detectorName?: string): PatternDetector | null {
    const detectorsToCheck = detectorName ? allDetectors.filter((d) => d.name === detectorName) : allDetectors;

    for (const detector of detectorsToCheck) {
        for (const regex of detector.regexes) {
            if (regex.test(message.toLowerCase())) {
                return detector;
            }
        }
    }
    return null;
}

// Generic response generation
export function generateResponse(detector: PatternDetector): string {
    const random = Math.random() * 100;
    let cumulativeProbability = 0;
    for (const response of detector.responses) {
        cumulativeProbability += response.probability;
        if (random <= cumulativeProbability) {
            return response.response;
        }
    }
    return detector.defaultResponse;
}
