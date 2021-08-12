export const name = 'mangadex-dl';

import * as API from "/static/js/api.js"
import * as MAPI from "/static/js/mangadex.js"
import * as Util from "/static/js/util.js"

function resultOk(resp) {
    return resp.result === "ok"
}

export default function (uid) {
    return {
        search: "",
        uid: uid,
        data: {},
        chapters: [],
        checkboxes: [],

        get filteredChapters() {
            return this.chapters.filter(
                i => {
                    let a = Util.Search.Match(this.search, i.title)
                    let b = Util.Search.Match(this.search, i.chapter)
                    let c = Util.Search.Match(this.search, i.volume)
                    let d = Util.Search.Match(this.search, i.scanlation_group)

                    return (a || b || c || d)
                }
            )
        },

        selectAll(e) {
            this.checkboxes.fill(e.target.checked)
        },

        async downloadChapters() {
            let chapters = []
            for (let i = 0; i < this.checkboxes.length; i++) {
                if (this.checkboxes[i] === true) {
                    chapters.push(this.chapters[i].raw)
                }
            }

            let data = {
                title: this.data.title,
                chapters: chapters,
            }
            await API.Download.Chapters(data)
        },

        async init() {
            await MAPI.Manga.View(this.uid, true)
                .then(resp => {
                    if (!resultOk(resp)) {
                        console.error("result not ok:", resp)
                        return
                    }

                    let d = {
                        id: resp.data.id,
                        createdAt: Util.Fmt.RFCDate(resp.data.attributes.createdAt),
                        title: Util.Ensure.String(resp.data.attributes.title.en),
                        description: Util.Ensure.String(resp.data.attributes.description.en),
                    }
                    for (let j = 0; j < resp.relationships.length; j++) {
                        let r = resp.relationships[j]
                        if (r.type === "cover_art") {
                            d.src = `https://uploads.mangadex.org/covers/${resp.data.id}/${r.attributes.fileName}.256.jpg`
                        }
                    }
                    this.data = d
                })

            await MAPI.Manga.FeedAll(this.uid)
                .then(resp => {
                    for (const r of resp) {
                        let d = {
                            id: r.data.id,
                            title: Util.Ensure.String(r.data.attributes.title),
                            volume: Util.Ensure.String(r.data.attributes.volume),
                            chapter: Util.Ensure.String(r.data.attributes.chapter),
                            updatedAt: Util.Fmt.RFCDate(r.data.attributes.updatedAt),
                            scanlation_group: r.relationships.find(o => { return o.type === "scanlation_group"}).attributes.name,
                            raw: r.data,
                        }

                        this.chapters.push(d)
                    }
                })

            this.checkboxes = new Array(this.chapters.length).fill(false)

            this.$watch('checkboxes', value => {
                let disabled = value.filter((i) => { return i === true }).length === 0
                document.getElementById('downloadButton').disabled = disabled
            })
        },
    }
}