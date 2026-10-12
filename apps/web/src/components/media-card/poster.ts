export const getPosterSrcSet = (src: string) => {
    if (!/^[^?#]*\/t\/p\/w500\/[^?#]+/.test(src)) {
        return undefined
    }

    return [185, 342, 500]
        .map((width) => `${src.replace("/t/p/w500/", `/t/p/w${width}/`)} ${width}w`)
        .join(", ")
}
