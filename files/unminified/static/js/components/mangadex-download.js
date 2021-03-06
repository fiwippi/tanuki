export const name = 'mangadex-dl';

import * as API from "/static/js/api.js"
import * as MAPI from "/static/js/mangadex.js"
import * as Util from "/static/js/util.js"

export default function (uid) {
    return {
        resp: "",
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
            console.log(chapters)
            await API.Download.Chapters(data)
                .then(() => {
                    this.resp = "Queued chapters"
                })
                .catch(() => {
                    this.resp = "Failed to queue chapters"
                })
        },

        async init() {
            document.getElementById("spinner").classList.add("loader")

            await MAPI.Manga.View(this.uid, true)
                .then(resp => {
                    let d = {
                        id: resp.id,
                        createdAt: Util.Fmt.RFCDate(resp.data.attributes.createdAt),
                        title: Util.Ensure.String(Object.values(resp.data.attributes.title)[0]),
                        description: Util.Ensure.String(resp.data.attributes.description.en),
                    }

                    for (let j = 0; j < resp.data.relationships.length; j++) {
                        let r = resp.data.relationships[j]
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
                            id: r.id,
                            title: Util.Ensure.String(r.attributes.title),
                            volume: Util.Ensure.String(r.attributes.volume),
                            chapter: Util.Ensure.String(r.attributes.chapter),
                            updatedAt: Util.Fmt.RFCDate(r.attributes.updatedAt),
                            scanlation_group: Util.Ensure.ScanlationGroup(r),
                            raw: r,
                        }

                        this.chapters.push(d)
                    }
                })

            this.checkboxes = new Array(this.chapters.length).fill(false)

            document.getElementById("spinner").classList.remove("loader")

            this.$watch('checkboxes', value => {
                let disabled = value.filter((i) => { return i === true }).length === 0
                document.getElementById('downloadButton').disabled = disabled
            })
        },
    }
}