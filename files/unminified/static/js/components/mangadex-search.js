export const name = 'mangadex-search';

import * as MAPI from "/static/js/mangadex.js"
import * as Util from "/static/js/util.js"

function resultOk(resp) {
    return resp.result === "ok"
}

export default function () {
    return {
        search: "",
        searchData: [],

        async searchEntries() {
            if (this.search.length === 0) {
                return []
            }

            await MAPI.Search.Manga(this.search, 8, true)
                .then(resp => {
                    for (let i = 0; i < resp.results.length; i++) {
                        let item = resp.results[i]
                        if (!resultOk(item)) {
                            console.error("result not ok:", item)
                            continue
                        }

                        let data = {
                            id: item.data.id,
                            createdAt: Util.Fmt.RFCDate(item.data.attributes.createdAt),
                            title: Object.values(item.data.attributes.title)[0],
                            description: item.data.attributes.description.en,
                        }
                        for (let j = 0; j < item.relationships.length; j++) {
                            let r = item.relationships[j]
                            if (r.type === "cover_art") {
                                data.src = `https://uploads.mangadex.org/covers/${item.data.id}/${r.attributes.fileName}.256.jpg`
                            }
                        }

                        this.searchData.push(data)
                    }
                })

            return this.searchData
        },

        fmtDescription(d) {
            if (d.length > 330) {
                return d.substring(0, 329) + "..."
            }
            return d
        },

        fmtDlLink(e) {
          return `/download/mangadex/${e.id}`
        },

        handleSearchChange(e) {
            this.searchData = []
        },
    }
}