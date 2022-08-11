export const name = 'specific-tag';

import createEntries from "/static/js/components/entries.js"

export default function (entries) {
    console.log(entries)

    let urlFunc = (i) => {
        return `/api/series/${entries[i].sid}/cover?thumbnail=true`
    }
    let idFunc = (i) => {
        return entries[i].sid
    }
    let extra = {
        tag: "",

        async preInit() {
            let prefix = "/tags/"
            this.tag = window.location.pathname.slice(prefix.length)
        },
    }

    return createEntries(entries, undefined, urlFunc, idFunc, extra)
}