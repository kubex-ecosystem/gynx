export const parseJsonResponse = <T>(responseText: string, schemaType: string): T => {
    try {
        if (!responseText) {
            throw new Error(`Received empty response from API when expecting ${schemaType}.`);
        }
        const cleanedJson = responseText.replace(/```json/g, '').replace(/```/g, '').trim();
        return JSON.parse(cleanedJson) as T;
    } catch (error) {
        console.error(`Error parsing ${schemaType} JSON:`, error);
        console.error("Invalid JSON string:", responseText);
        throw new Error(`Failed to parse the analysis response. The format was invalid.`);
    }
};

export const handleGeminiError = (error: any) => {
    console.error("Gemini API Error:", error);
    const message = error.toString();

    if (message.includes("API_KEY_EMPTY")) {
        throw new Error("The API key field cannot be empty. Please enter a valid key.");
    }
    if (message.includes("API key not valid")) {
        throw new Error("The Gemini API key is invalid. Please check your key and try again.");
    }
    if (message.includes("Billing") || message.includes("billing")) {
        throw new Error("There seems to be a billing issue with your Google Cloud project for the Gemini API.");
    }
    if (message.includes("SAFETY")) {
        throw new Error("The response was blocked due to safety settings. Please modify your input.");
    }

    throw new Error("An unexpected error occurred with the Gemini API.");
};