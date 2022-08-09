export const name = 'catalog';

import * as Util from "/static/js/util.js"
import createEntries from "/static/js/components/entries.js"

export default function (entries, progress) {
    console.debug("catalog entries:", entries)
    console.debug("catalog progress:", progress)

    let urlFunc = (i) => {
        return `/api/series/${entries[i].sid}/cover?thumbnail=true`
    }
    let idFunc = (i) => {
        return entries[i].sid
    }
    let extra = {
        fmtProgress(e) {
            let prog = this.progress[e.sid]
            let pages = e.num_pages
            return `Progress: ${Util.Fmt.SeriesPercent(prog, pages)}`
        }
    }

    return createEntries(entries, progress, urlFunc, idFunc, extra)
}