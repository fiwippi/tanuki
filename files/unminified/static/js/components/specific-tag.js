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

            // We need to reindex the entries starting at 1 and
            // moving up so that the thumbnails can be retrieved
            // with the getThumbnail() function
            for (let i = 0; i < this.entries.length; i++) {
                this.entries[i].order = i + 1;
            }
        },
    }

    return createEntries(entries, undefined, urlFunc, extra)
}