export const name = 'catalog';

import * as Util from "/static/js/util.js"
import createEntries from "/static/js/components/entries.js"

export default function (entries, progress) {
    console.debug("catalog progress:", progress)

    let urlFunc = (i) => {
        return `/api/series/${entries[i].hash}/cover?thumbnail=true`
    }
    let extra = {
        fmtProgress(e) {
            let prog = this.progress[e.hash]
            let pages = e.total_pages
            return `Progress: ${Util.Fmt.SeriesPercent(prog, pages)}`
        }
    }

    return createEntries(entries, progress, urlFunc, extra)
}