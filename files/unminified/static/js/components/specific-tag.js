export const name = 'specific-tag';

import createEntries from "/static/js/components/entries.js"

export default function (entries) {
    let urlFunc = (i) => {
        return `/api/series/${entries[i].hash}/cover?thumbnail=true`
    }
    let extra = {
        tag: "",

        async preInit() {
            let prefix = "/tags/"
            this.tag = window.location.pathname.slice(prefix.length)
        },
    }

    return createEntries(entries, undefined, urlFunc, extra)
}