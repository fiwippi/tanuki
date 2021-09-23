export const name = 'mangadex-search';

import * as MAPI from "/static/js/mangadex.js"
import * as Util from "/static/js/util.js"

export default function () {
    return {
        search: "",
        searchData: [],

        async searchEntries() {
            if (this.search.length === 0) {
                return []
            }

            document.getElementById("spinner").classList.add("loader")

            await MAPI.Search.Manga(this.search, 8, true)
                .then(resp => {
                    if (resp.data === undefined) {
                        console.debug(resp)
                        return
                    }
                    for (let i = 0; i < resp.data.length; i++) {
                        let item = resp.data[i]

                        let data = {
                            id: item.id,
                            createdAt: Util.Fmt.RFCDate(item.attributes.createdAt),
                            title: Object.values(item.attributes.title)[0],
                            description: item.attributes.description.en,
                        }
                        for (let j = 0; j < item.relationships.length; j++) {
                            let r = item.relationships[j]
                            if (r.type === "cover_art") {
                                data.src = `https://uploads.mangadex.org/covers/${item.id}/${r.attributes.fileName}.256.jpg`
                            }
                        }

                        this.searchData.push(data)
                    }
                })

            document.getElementById("spinner").classList.remove("loader")

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