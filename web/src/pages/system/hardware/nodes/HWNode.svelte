<script lang="ts">
    import IconButton from "@smui/icon-button";
    import type { hardwareNode } from "../types";

    export let data: hardwareNode = {
        url: "",
        name: "",
        token: "",
        enabled: false,
        online: false,
    };
</script>

<div class="node mdc-elevation--z3">
    <div class="node__header">
        <span class="node__header__name">{data.name}</span>
        <span class="node__header__url text-hint">{data.url}</span>
    </div>
    <div class="node__footer">
        <IconButton class="material-icons">edit</IconButton>
        <IconButton class="material-icons">delete</IconButton>
    </div>
</div>

<style lang="scss">
    @use "../../../../mixins" as *;

    .node {
        background-color: var(--clr-height-1-3);
        position: relative;
        border-radius: 0.3rem;
        padding: 1rem;
        height: 6rem;
        width: 14rem;

        @include widescreen {
            width: 16rem;
        }

        &__header {
            &__name {
                display: block;
            }
        }
        &__footer {
            position: absolute;
            right: 10px;
            bottom: 5px;
        }
    }
</style>
