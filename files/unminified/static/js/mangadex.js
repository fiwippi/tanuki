export const name = 'mangadex';

const API_URL = "https://api.mangadex.org/"

async function fetchResource(route, userOptions = {}) {
    // Define default options
    const defaultOptions = {
        method: 'GET',
    };
    // Define default headers
    const defaultHeaders = {
        'Content-Type': 'application/json',
    };

    const options = {
        // Merge options
        ...defaultOptions,
        ...userOptions,
        // Merge headers
        headers: {
            ...defaultHeaders,
            ...userOptions.headers,
        },
    };

    return fetch(API_URL + route, options)
        .then(response => {
            if (response.status >= 400 && response.status <= 500) {
                throw Error(`invalid response status: ${response.status}, resp: ${response.body}`)
                return
            }
            return response.json()
        })
        .catch(error => {throw error})
}

export class Search {
    static async Manga(title, limit, coverArt) {
        if (limit === undefined) limit = 15
        return fetchResource(`manga?title=${title}&limit=${limit}${coverArt ? "&includes[]=cover_art" : ""}`)
    }
}

export class Manga {
    static async View(id, coverArt) {
        return fetchResource(`manga/${id}${coverArt ? "?includes[]=cover_art" : ""}`)
    }

    static async Feed(id, limit, offset) {
        if (limit === undefined) limit = 10
        if (offset === undefined) offset = 0
        return fetchResource(`manga/${id}/feed?limit=${limit}&offset=${offset}&translatedLanguage[]=en&order[chapter]=desc&includes[]=scanlation_group`)
    }

    static async FeedAll(id) {
        let feed = []
        let moreLeft = true
        let limit = 500
        let offset = 0

        while (moreLeft) {
            await this.Feed(id, limit, offset)
                .then(resp => {
                    feed.push(...resp.results)
                    offset += limit
                    if (feed.length >= resp.total) {
                        moreLeft = false
                    }
                })
        }

        return feed
    }

    static async Chapters(id) {
        return fetchResource(`manga/${id}/aggregate`)
    }
}