export const name = 'catalog';

import * as Util from "/static/js/util.js"
import createEntries from "/static/js/components/entries.js"

export default function (entries, progress) {
    let urlFunc = (i) => {
        return `/api/series/${entries[i].hash}/cover?thumbnail=true`
    }
    let extra = {
        fmtProgress(p) {
            return `Progress: ${Util.Fmt.SeriesPercent(p)}`
        }
    }

    return createEntries(entries, progress, urlFunc, extra)
}