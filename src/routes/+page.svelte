<script lang="ts">
    import { onMount } from "svelte";
    import { invoke } from "@tauri-apps/api/core";
    import Chart from "chart.js/auto";
    import { save } from "@tauri-apps/plugin-dialog";
    import { writeTextFile } from "@tauri-apps/plugin-fs";

    // ── State ────────────────────────────────────────────────
    let records = $state([]);
    let dateFrom = $state("");
    let dateTo = $state("");
    let activePeriod = $state("3m");

    // add-expense form
    let eVal = $state("");
    let eDesc = $state("");
    let eCat = $state("");
    let eDate = $state("");

    // add-revenue form
    let rVal = $state("");
    let rDesc = $state("");
    let rCat = $state("");
    let rDate = $state("");

    // dynamic category lists derived from all records ever loaded
    let expCategories = $derived(
        [
            ...new Set(
                records.filter((r) => r.type === "expense").map((r) => r.cat),
            ),
        ].sort(),
    );
    let revCategories = $derived(
        [
            ...new Set(
                records.filter((r) => r.type === "revenue").map((r) => r.cat),
            ),
        ].sort(),
    );

    // ── Derived period data ───────────────────────────────────
    let filtered = $derived(
        records.filter((r) => r.date >= dateFrom && r.date <= dateTo),
    );
    let expenses = $derived(filtered.filter((r) => r.type === "expense"));
    let revenues = $derived(filtered.filter((r) => r.type === "revenue"));
    let totalExp = $derived(expenses.reduce((s, r) => s + r.val, 0));
    let totalRev = $derived(revenues.reduce((s, r) => s + r.val, 0));
    let balance = $derived(totalRev - totalExp);
    let last4 = $derived(
        [...records].sort((a, b) => b.date.localeCompare(a.date)).slice(0, 4),
    );
    let ledgerSorted = $derived(
        [...filtered].sort((a, b) => b.date.localeCompare(a.date)),
    );

    // ── Variation: compare current period to equally-sized previous period ──
    let varExp = $derived.by(() => {
        if (!dateFrom || !dateTo) return 0;
        const ms = new Date(dateTo) - new Date(dateFrom);
        const prevTo = new Date(new Date(dateFrom) - 1)
            .toISOString()
            .split("T")[0];
        const prevFrom = new Date(new Date(dateFrom) - ms - 1)
            .toISOString()
            .split("T")[0];
        const prevExp = records
            .filter(
                (r) =>
                    r.type === "expense" &&
                    r.date >= prevFrom &&
                    r.date <= prevTo,
            )
            .reduce((s, r) => s + r.val, 0);
        if (prevExp === 0) return null;
        return ((totalExp - prevExp) / prevExp) * 100;
    });

    let varRev = $derived.by(() => {
        if (!dateFrom || !dateTo) return 0;
        const ms = new Date(dateTo) - new Date(dateFrom);
        const prevTo = new Date(new Date(dateFrom) - 1)
            .toISOString()
            .split("T")[0];
        const prevFrom = new Date(new Date(dateFrom) - ms - 1)
            .toISOString()
            .split("T")[0];
        const prevRev = records
            .filter(
                (r) =>
                    r.type === "revenue" &&
                    r.date >= prevFrom &&
                    r.date <= prevTo,
            )
            .reduce((s, r) => s + r.val, 0);
        if (prevRev === 0) return null;
        return ((totalRev - prevRev) / prevRev) * 100;
    });

    // --- Exporting -----------------------
    async function exportJSON() {
        const data = await invoke("export_json");
        const path = await save({
            filters: [{ name: "JSON", extensions: ["json"] }],
        });
        if (path) await writeTextFile(path, data);
    }

    async function exportCSV() {
        const data = await invoke("export_csv");
        const path = await save({
            filters: [{ name: "CSV", extensions: ["csv"] }],
        });
        if (path) await writeTextFile(path, data);
    }

    // ── Category breakdown ────────────────────────────────────
    function catBreakdown(items, total) {
        const m = {};
        items.forEach((r) => {
            m[r.cat] = (m[r.cat] || 0) + r.val;
        });
        return Object.entries(m)
            .sort((a, b) => b[1] - a[1])
            .map(([cat, val]) => ({
                cat,
                val,
                pct: total > 0 ? ((val / total) * 100).toFixed(1) : "0.0",
            }));
    }
    let expCats = $derived(catBreakdown(expenses, totalExp));
    let revCats = $derived(catBreakdown(revenues, totalRev));

    // ── Formatting ────────────────────────────────────────────
    const MO = [
        "Jan",
        "Feb",
        "Mar",
        "Apr",
        "May",
        "Jun",
        "Jul",
        "Aug",
        "Sep",
        "Oct",
        "Nov",
        "Dec",
    ];
    const brl = (v) =>
        "R$ " +
        v
            .toFixed(2)
            .replace(".", ",")
            .replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    const fdt = (d) => {
        const [y, m, day] = d.split("-");
        return `${day} ${MO[+m - 1]} ${y}`;
    };
    const fmtVar = (v) =>
        v === null ? "n/a" : (v >= 0 ? "+" : "") + v.toFixed(1) + "%";

    // ── Period control ────────────────────────────────────────
    function setPeriod(p) {
        activePeriod = p;
        const now = new Date();
        const from = new Date(now);
        if (p === "1m") from.setMonth(from.getMonth() - 1);
        else if (p === "3m") from.setMonth(from.getMonth() - 3);
        else if (p === "6m") from.setMonth(from.getMonth() - 6);
        else if (p === "1y") from.setFullYear(from.getFullYear() - 1);
        dateFrom = from.toISOString().split("T")[0];
        dateTo = now.toISOString().split("T")[0];
        loadRecords();
    }

    function applyDates() {
        activePeriod = "";
        loadRecords();
    }

    // ── Backend calls ─────────────────────────────────────────
    async function loadRecords() {
        try {
            records = await invoke("get_transactions", {
                from: dateFrom,
                to: dateTo,
            });
        } catch (e) {
            console.error("get_transactions failed:", e);
        }
    }

    async function addEntry(type) {
        const isExp = type === "expense";
        const val = parseFloat(isExp ? eVal : rVal);
        const desc = isExp ? eDesc.trim() : rDesc.trim();
        const cat = isExp ? eCat.trim() : rCat.trim();
        const date = isExp ? eDate : rDate;
        if (!desc || isNaN(val) || val <= 0 || !date || !cat) return;

        try {
            await invoke("add_transaction", {
                transaction: { type, desc, cat, val, date },
            });
            if (isExp) {
                eVal = "";
                eDesc = "";
            } else {
                rVal = "";
                rDesc = "";
            }
            await loadRecords();
        } catch (e) {
            console.error("add_transaction failed:", e);
        }
    }

    // ── Charts ────────────────────────────────────────────────
    const chartOpts = (color) => ({
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                labels: {
                    font: { family: "'Caveat', cursive", size: 13 },
                    color: "#4a3c2e",
                    boxWidth: 10,
                },
            },
            tooltip: {
                backgroundColor: "#1c140d",
                titleColor: "#f4eed8",
                bodyColor: "#c8bda6",
                titleFont: { family: "'Caveat', cursive", size: 14 },
                bodyFont: { family: "'Caveat', cursive", size: 13 },
                callbacks: {
                    label: (c) =>
                        " " + c.dataset.label + ": " + brl(c.parsed.y),
                },
            },
        },
        scales: {
            x: {
                ticks: {
                    font: { family: "'Caveat', cursive", size: 12 },
                    color: "#8a7a66",
                },
                grid: { color: "#e5ddc8" },
                border: { color: "#c8bda6" },
            },
            y: {
                ticks: {
                    font: { family: "'Caveat', cursive", size: 12 },
                    color: "#8a7a66",
                    callback: (v) => "R$" + v.toLocaleString("pt-BR"),
                },
                grid: { color: "#e5ddc8" },
                border: { color: "#c8bda6" },
            },
        },
    });

    function bucketByMonth(items) {
        const b = {};
        items.forEach((r) => {
            const k = r.date.slice(0, 7);
            b[k] = (b[k] || 0) + r.val;
        });
        const ks = Object.keys(b).sort();
        return {
            ks,
            labels: ks.map((k) => {
                const [y, m] = k.split("-");
                return MO[+m - 1] + " " + y.slice(2);
            }),
            vals: ks.map((k) => b[k]),
        };
    }

    // canvas refs
    let lineEl, donutEl, expLineEl, revLineEl;
    let lc, dc, elc, rlc;

    $effect(() => {
        // main line chart
        const { ks, labels } = (() => {
            const b = {};
            filtered.forEach((r) => {
                const k = r.date.slice(0, 7);
                if (!b[k]) b[k] = { e: 0, r: 0 };
                b[k][r.type === "expense" ? "e" : "r"] += r.val;
            });
            const ks = Object.keys(b).sort();
            return {
                ks,
                b,
                labels: ks.map((k) => {
                    const [y, m] = k.split("-");
                    return MO[+m - 1] + " " + y.slice(2);
                }),
                b,
            };
        })();
        const bm = (() => {
            const b = {};
            filtered.forEach((r) => {
                const k = r.date.slice(0, 7);
                if (!b[k]) b[k] = { e: 0, r: 0 };
                b[k][r.type === "expense" ? "e" : "r"] += r.val;
            });
            return b;
        })();
        if (lc) lc.destroy();
        if (lineEl)
            lc = new Chart(lineEl, {
                type: "line",
                data: {
                    labels,
                    datasets: [
                        {
                            label: "Expenses",
                            data: ks.map((k) => bm[k]?.e || 0),
                            borderColor: "#8c1f1f",
                            backgroundColor: "rgba(140,31,31,.07)",
                            borderWidth: 1.5,
                            pointRadius: 4,
                            pointBackgroundColor: "#8c1f1f",
                            tension: 0.4,
                            fill: true,
                        },
                        {
                            label: "Revenue",
                            data: ks.map((k) => bm[k]?.r || 0),
                            borderColor: "#1a5c30",
                            backgroundColor: "rgba(26,92,48,.07)",
                            borderWidth: 1.5,
                            pointRadius: 4,
                            pointBackgroundColor: "#1a5c30",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        // donut
        const pal = [
            "#8c1f1f",
            "#b84040",
            "#c97060",
            "#d49880",
            "#a05030",
            "#7a3010",
            "#5a2008",
        ];
        const em = {};
        expenses.forEach((r) => {
            em[r.cat] = (em[r.cat] || 0) + r.val;
        });
        if (dc) dc.destroy();
        if (donutEl)
            dc = new Chart(donutEl, {
                type: "doughnut",
                data: {
                    labels: Object.keys(em),
                    datasets: [
                        {
                            data: Object.values(em),
                            backgroundColor: pal.slice(
                                0,
                                Object.keys(em).length,
                            ),
                            borderColor: "#f4eed8",
                            borderWidth: 2,
                        },
                    ],
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    cutout: "76%",
                    plugins: {
                        legend: {
                            position: "right",
                            labels: {
                                font: { family: "'Caveat',cursive", size: 12 },
                                color: "#4a3c2e",
                                boxWidth: 10,
                                padding: 8,
                            },
                        },
                        tooltip: {
                            backgroundColor: "#1c140d",
                            titleColor: "#f4eed8",
                            bodyColor: "#c8bda6",
                            titleFont: { family: "'Caveat',cursive", size: 13 },
                            bodyFont: { family: "'Caveat',cursive", size: 12 },
                            callbacks: {
                                label: (c) => {
                                    const p =
                                        totalExp > 0
                                            ? (
                                                  (c.parsed / totalExp) *
                                                  100
                                              ).toFixed(1)
                                            : 0;
                                    return (
                                        " " + brl(c.parsed) + " (" + p + "%)"
                                    );
                                },
                            },
                        },
                    },
                },
            });

        // expenses line
        const eb = bucketByMonth(expenses);
        if (elc) elc.destroy();
        if (expLineEl)
            elc = new Chart(expLineEl, {
                type: "line",
                data: {
                    labels: eb.labels,
                    datasets: [
                        {
                            label: "Expenses",
                            data: eb.vals,
                            borderColor: "#8c1f1f",
                            backgroundColor: "rgba(140,31,31,.07)",
                            borderWidth: 1.5,
                            pointRadius: 4,
                            pointBackgroundColor: "#8c1f1f",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        // revenues line
        const rb = bucketByMonth(revenues);
        if (rlc) rlc.destroy();
        if (revLineEl)
            rlc = new Chart(revLineEl, {
                type: "line",
                data: {
                    labels: rb.labels,
                    datasets: [
                        {
                            label: "Revenue",
                            data: rb.vals,
                            borderColor: "#1a5c30",
                            backgroundColor: "rgba(26,92,48,.07)",
                            borderWidth: 1.5,
                            pointRadius: 4,
                            pointBackgroundColor: "#1a5c30",
                            tension: 0.4,
                            fill: true,
                        },
                    ],
                },
                options: chartOpts(),
            });

        return () => {
            lc?.destroy();
            dc?.destroy();
            elc?.destroy();
            rlc?.destroy();
        };
    });

    // ── Init ──────────────────────────────────────────────────
    onMount(() => {
        const today = new Date().toISOString().split("T")[0];
        eDate = today;
        rDate = today;
        setPeriod("3m");
    });
</script>

<!-- ══════════════════════════════════════════════════════════ TOPBAR -->
<header id="topbar">
    <div class="brand">Self Ledger</div>

    <div class="period-ctrl">
        {#each [["1m", "1m"], ["3m", "3m"], ["6m", "6m"], ["1y", "1y"]] as [p, label]}
            <button
                class="pbtn"
                class:on={activePeriod === p}
                onclick={() => setPeriod(p)}>{label}</button
            >
        {/each}
        <span class="psep">|</span>
        <input
            type="date"
            class="dinput"
            bind:value={dateFrom}
            onchange={applyDates}
        />
        <span style="color:var(--ink-faint);padding:0 4px">→</span>
        <input
            type="date"
            class="dinput"
            bind:value={dateTo}
            onchange={applyDates}
        />
    </div>

    <div class="nav-right">
        <div class="nav-totals">
            <div class="nav-tot">
                <span class="nav-tot-label">Expenses</span>
                <span class="nav-tot-val exp">{brl(totalExp)}</span>
            </div>
            <div class="nav-tot">
                <span class="nav-tot-label">Revenues</span>
                <span class="nav-tot-val rev">{brl(totalRev)}</span>
            </div>
            <div class="nav-tot">
                <span class="nav-tot-label">Balance</span>
                <span
                    class="nav-tot-val"
                    class:pos={balance >= 0}
                    class:neg={balance < 0}>{brl(Math.abs(balance))}</span
                >
            </div>
        </div>

        <div class="nav-exports">
            <button class="exp-btn" onclick={exportJSON}>↓ JSON</button>
            <button class="exp-btn" onclick={exportCSV}>↓ CSV</button>
        </div>
    </div>
</header>

<!-- ══════════════════════════════════════════════════════════ PAGE -->
<div id="wrap">
    <div id="inner">
        <!-- ══ SECTION 1 — Last 4 records | Line chart ══ -->
        <div class="section wide-left">
            <div class="vcol">
                <div>
                    {#each last4 as r}
                        <div class="row row-3">
                            <span
                                class="val"
                                class:exp={r.type === "expense"}
                                class:rev={r.type === "revenue"}
                                >{brl(r.val)}</span
                            >
                            <span class="meta">{r.desc} · {r.cat}</span>
                            <span class="dt">{fdt(r.date)}</span>
                        </div>
                    {/each}
                </div>

                <div style="height:20px"></div>

                <!-- ADD EXPENSE -->
                <div class="add-row">
                    <input
                        class="ii v exp-ii"
                        type="number"
                        placeholder="R$ 0,00"
                        step=".01"
                        min="0"
                        bind:value={eVal}
                    />
                    <input
                        class="ii d"
                        type="text"
                        placeholder="description"
                        bind:value={eDesc}
                    />
                    <div class="cat-combo">
                        <input
                            class="ii cat-ii exp-ii"
                            type="text"
                            placeholder="category"
                            list="exp-cats-list"
                            bind:value={eCat}
                        />
                        <datalist id="exp-cats-list">
                            {#each expCategories as c}<option
                                    value={c}
                                />{/each}
                        </datalist>
                    </div>
                    <input class="ii t" type="date" bind:value={eDate} />
                    <button
                        class="abtn exp-abtn"
                        onclick={() => addEntry("expense")}>↵</button
                    >
                </div>

                <!-- ADD REVENUE -->
                <div class="add-row">
                    <input
                        class="ii v rev-ii"
                        type="number"
                        placeholder="R$ 0,00"
                        step=".01"
                        min="0"
                        bind:value={rVal}
                    />
                    <input
                        class="ii d"
                        type="text"
                        placeholder="description"
                        bind:value={rDesc}
                    />
                    <div class="cat-combo">
                        <input
                            class="ii cat-ii rev-ii"
                            type="text"
                            placeholder="category"
                            list="rev-cats-list"
                            bind:value={rCat}
                        />
                        <datalist id="rev-cats-list">
                            {#each revCategories as c}<option
                                    value={c}
                                />{/each}
                        </datalist>
                    </div>
                    <input class="ii t" type="date" bind:value={rDate} />
                    <button
                        class="abtn rev-abtn"
                        onclick={() => addEntry("revenue")}>↵</button
                    >
                </div>
            </div>
            <div class="vdiv"></div>
            <div class="vcol">
                <div class="chart-wrap h200">
                    <canvas bind:this={lineEl}></canvas>
                </div>
            </div>
        </div>

        <!-- ══ SECTION 2 — Donut | Expense category list ══ -->
        <div class="section">
            <div class="vcol">
                <div class="chart-label">expenses per category</div>
                <div class="chart-wrap h160">
                    <canvas bind:this={donutEl}></canvas>
                </div>
            </div>
            <div class="vdiv"></div>
            <div class="vcol">
                {#each expCats as { cat, pct }}
                    <div class="cat-row">
                        <span class="cat-name">{cat}</span>
                        <span class="cat-pct exp">{pct}%</span>
                    </div>
                {/each}
            </div>
        </div>

        <!-- ══ SECTION 3 — Expenses ══ -->
        <div class="block-section">
            <div class="block-title exp">expenses</div>
            <div class="block-top">
                <div>
                    <div class="chart-wrap h200">
                        <canvas bind:this={expLineEl}></canvas>
                    </div>
                </div>
                <div class="annot" style="margin-top:0;">
                    <div class="annot-label">notes</div>
                    <textarea rows="8" placeholder="personal annotations..."
                    ></textarea>
                </div>
            </div>
            <div class="block-cats">
                <div class="sec-label" style="margin-bottom:8px;">
                    by category
                </div>
                <div class="block-cats-grid">
                    {#each expCats as { cat, pct }}
                        <div class="cat-row">
                            <span class="cat-name">{cat}</span>
                            <span class="cat-pct exp">{pct}%</span>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <!-- ══ SECTION 4 — Revenues ══ -->
        <div class="block-section">
            <div class="block-title rev">revenues</div>
            <div class="block-top">
                <div>
                    <div class="chart-wrap h200">
                        <canvas bind:this={revLineEl}></canvas>
                    </div>
                </div>
                <div class="annot" style="margin-top:0;">
                    <div class="annot-label">notes</div>
                    <textarea rows="8" placeholder="personal annotations..."
                    ></textarea>
                </div>
            </div>
            <div class="block-cats">
                <div class="sec-label" style="margin-bottom:8px;">
                    by category
                </div>
                <div class="block-cats-grid">
                    {#each revCats as { cat, pct }}
                        <div class="cat-row">
                            <span class="cat-name">{cat}</span>
                            <span class="cat-pct rev">{pct}%</span>
                        </div>
                    {/each}
                </div>
            </div>
        </div>

        <!-- ══ SECTION 5 — Full ledger | Variation ══ -->
        <div class="section wide-left">
            <div class="vcol">
                <div class="drule"></div>
                <div class="sec-label" style="margin-top:10px;">ledger</div>
                {#each ledgerSorted as r}
                    <div class="row row-3">
                        <span
                            class="val"
                            class:exp={r.type === "expense"}
                            class:rev={r.type === "revenue"}>{brl(r.val)}</span
                        >
                        <span class="meta">{r.desc} · {r.cat}</span>
                        <span class="dt">{fdt(r.date)}</span>
                    </div>
                {/each}
            </div>
            <div class="vdiv"></div>
            <div
                class="vcol"
                style="display:flex;flex-direction:column;justify-content:center;"
            >
                <div class="var-row">
                    <span class="var-letter rev">R</span>
                    <span
                        class="var-val"
                        class:up={varRev !== null && varRev >= 0}
                        class:dn={varRev !== null && varRev < 0}
                        >{fmtVar(varRev)}</span
                    >
                </div>
                <div class="var-row" style="margin-top:8px;">
                    <span class="var-letter exp">E</span>
                    <span
                        class="var-val"
                        class:dn={varExp !== null && varExp > 0}
                        class:up={varExp !== null && varExp <= 0}
                        >{fmtVar(varExp)}</span
                    >
                </div>
            </div>
        </div>
    </div>
</div>

<style>
    :global(:root) {
        --paper: #f4eed8;
        --ink: #1c140d;
        --ink-mid: #4a3c2e;
        --ink-faint: #8a7a66;
        --rule: #c8bda6;
        --rule-dark: #9a8a74;
        --red: #8c1f1f;
        --green: #1a5c30;
    }
    :global(*) {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }
    :global(body) {
        background: var(--paper);
        color: var(--ink);
        font-family: "Caveat", cursive;
        font-size: 17px;
        line-height: 1.5;
    }
    :global(body::after) {
        content: "";
        position: fixed;
        inset: 0;
        background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='300' height='300'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.7' numOctaves='4' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='300' height='300' filter='url(%23n)' opacity='0.04'/%3E%3C/svg%3E");
        pointer-events: none;
        z-index: 999;
    }

    /* ── TOPBAR ── */
    #topbar {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        z-index: 100;
        background: var(--paper);
        padding: 8px 48px;
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        align-items: center;
        border-bottom: 2px solid var(--ink-mid);
        min-height: 64px;
    }
    .brand {
        font-family: "Caveat Brush", cursive;
        font-size: 24px;
        color: var(--ink);
    }
    .period-ctrl {
        display: flex;
        align-items: baseline;
        gap: 4px;
        justify-content: center;
    }
    .pbtn {
        background: none;
        border: none;
        border-bottom: 2px solid transparent;
        font-family: "Caveat", cursive;
        font-size: 16px;
        color: var(--ink-faint);
        cursor: pointer;
        padding: 2px 8px;
        transition: all 0.15s;
    }
    .pbtn:hover {
        color: var(--ink);
    }
    .pbtn.on {
        color: var(--ink);
        border-bottom-color: var(--ink);
    }
    .psep {
        color: var(--rule);
        padding: 0 6px;
    }
    .dinput {
        background: none;
        border: none;
        border-bottom: 1px solid var(--rule-dark);
        font-family: "Caveat", cursive;
        font-size: 15px;
        color: var(--ink-mid);
        padding: 0 4px;
        width: 108px;
        outline: none;
    }
    .dinput:focus {
        border-bottom-color: var(--ink);
    }
    .nav-right {
        display: flex;
        align-items: center;
        gap: 20px;
        justify-self: end;
    }
    .nav-totals {
        display: flex;
        gap: 24px;
        align-items: center;
    }
    .nav-tot {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
    }
    .nav-tot-label {
        font-family: "Caveat Brush", cursive;
        font-size: 10px;
        letter-spacing: 0.18em;
        text-transform: uppercase;
        color: var(--ink-faint);
        line-height: 1;
    }
    .nav-tot-val {
        font-size: 20px;
        font-weight: 700;
        line-height: 1.2;
    }
    .nav-tot-val.exp {
        color: var(--red);
    }
    .nav-tot-val.rev {
        color: var(--green);
    }
    .nav-tot-val.pos {
        color: var(--green);
    }
    .nav-tot-val.neg {
        color: var(--red);
    }
    .nav-exports {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 5px;
    }
    .exp-btn {
        font-family: "Caveat Brush", cursive;
        font-size: 13px;
        letter-spacing: 0.03em;
        background: none;
        border: 1.5px solid var(--ink-mid);
        color: var(--ink-mid);
        border-radius: 4px;
        padding: 2px 10px;
        cursor: pointer;
        line-height: 1.4;
        transition: background 0.15s, color 0.15s, border-color 0.15s;
        width: 76px;
        text-align: center;
    }
    .exp-btn:hover {
        background: var(--ink-mid);
        color: var(--paper);
        border-color: var(--ink-mid);
    }

    /* ── LAYOUT ── */
    #wrap {
        margin-top: 64px;
    }
    #inner {
        max-width: 1100px;
        margin: 0 auto;
        padding: 0 32px;
    }
    .section {
        display: grid;
        grid-template-columns: 1fr 2px 1fr;
    }
    .section.wide-left {
        grid-template-columns: 60fr 2px 40fr;
    }
    .vcol {
        padding: 28px 24px;
    }
    .vdiv {
        width: 2px;
    }

    /* ── TITLES ── */
    .sec-label {
        font-family: "Caveat Brush", cursive;
        font-size: 13px;
        letter-spacing: 0.2em;
        text-transform: uppercase;
        color: var(--ink-faint);
        margin-bottom: 10px;
    }
    .block-title {
        font-family: "Caveat Brush", cursive;
        font-size: 28px;
        padding-bottom: 6px;
        margin-bottom: 12px;
        border-bottom: 1.5px solid currentColor;
    }
    .block-title.exp {
        color: var(--red);
    }
    .block-title.rev {
        color: var(--green);
    }

    /* ── ROWS ── */
    .row {
        display: grid;
        align-items: baseline;
        padding: 5px 0;
        border-bottom: 1px solid var(--rule);
        gap: 10px;
    }
    .row-3 {
        grid-template-columns: 120px 1fr auto;
    }
    .val {
        font-size: 19px;
        font-weight: 700;
    }
    .val.exp {
        color: var(--red);
    }
    .val.rev {
        color: var(--green);
    }
    .meta {
        font-size: 16px;
        color: var(--ink-mid);
    }
    .dt {
        font-size: 14px;
        color: var(--ink-faint);
        white-space: nowrap;
        text-align: right;
    }

    /* ── ADD ROWS ── */
    .add-row {
        display: flex;
        align-items: baseline;
        gap: 8px;
        padding: 6px 0;
        border-bottom: 1px solid var(--rule);
    }
    .ii {
        background: none;
        border: none;
        border-bottom: 1px dashed var(--rule);
        font-family: "Caveat", cursive;
        font-size: 16px;
        color: var(--ink);
        outline: none;
        padding: 0 4px;
    }
    .ii:focus {
        border-bottom-color: var(--ink-mid);
    }
    .ii.v {
        width: 88px;
    }
    .ii.d {
        flex: 1;
    }
    .ii.t {
        width: 110px;
    }
    .exp-ii {
        color: var(--red);
    }
    .exp-ii::placeholder {
        color: var(--red);
        opacity: 0.6;
    }
    .rev-ii {
        color: var(--green);
    }
    .rev-ii::placeholder {
        color: var(--green);
        opacity: 0.6;
    }

    /* ── COMBOBOX CATEGORY ── */
    .cat-combo {
        position: relative;
        width: 120px;
    }
    .cat-ii {
        width: 100%;
    }

    .abtn {
        background: none;
        border: none;
        font-family: "Caveat", cursive;
        font-size: 17px;
        cursor: pointer;
        transition: color 0.15s;
        padding: 0 4px;
    }
    .exp-abtn {
        color: var(--red);
        opacity: 0.7;
    }
    .exp-abtn:hover {
        opacity: 1;
    }
    .rev-abtn {
        color: var(--green);
        opacity: 0.7;
    }
    .rev-abtn:hover {
        opacity: 1;
    }

    /* ── BLOCK SECTIONS ── */
    .block-section {
        padding: 28px 24px;
    }
    .block-top {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 32px;
        margin-bottom: 20px;
    }
    .block-cats {
        border-top: 1px solid var(--rule);
        padding-top: 16px;
    }
    .block-cats-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
        gap: 0 24px;
    }

    /* ── CATEGORY ROWS ── */
    .cat-row {
        display: grid;
        grid-template-columns: 1fr auto;
        align-items: baseline;
        padding: 4px 0;
        border-bottom: 1px solid var(--rule);
        gap: 10px;
    }
    .cat-name {
        font-size: 16px;
        color: var(--ink-mid);
    }
    .cat-pct {
        font-size: 17px;
        font-weight: 700;
    }
    .cat-pct.exp {
        color: var(--red);
    }
    .cat-pct.rev {
        color: var(--green);
    }

    /* ── CHARTS ── */
    .chart-wrap {
        width: 100%;
        position: relative;
    }
    .chart-wrap.h200 {
        height: 200px;
    }
    .chart-wrap.h160 {
        height: 160px;
    }
    .chart-label {
        font-family: "Caveat Brush", cursive;
        font-size: 14px;
        letter-spacing: 0.08em;
        color: var(--ink-faint);
        margin-bottom: 8px;
    }

    /* ── ANNOTATION ── */
    .annot {
        margin-top: 14px;
    }
    .annot-label {
        font-family: "Caveat Brush", cursive;
        font-size: 12px;
        letter-spacing: 0.15em;
        text-transform: uppercase;
        color: var(--ink-faint);
    }
    .annot textarea {
        display: block;
        width: 100%;
        background: none;
        border: none;
        border-bottom: 1px dashed var(--rule);
        font-family: "Caveat", cursive;
        font-size: 16px;
        font-style: italic;
        color: var(--ink-mid);
        resize: none;
        outline: none;
        line-height: 2;
        padding: 2px 0;
        margin-top: 2px;
    }

    /* ── VARIATION ── */
    .var-row {
        display: grid;
        grid-template-columns: 36px 1fr;
        align-items: baseline;
        padding: 6px 0;
        border-bottom: 1px solid var(--rule);
        gap: 8px;
    }
    .var-letter {
        font-family: "Caveat Brush", cursive;
        font-size: 26px;
    }
    .var-letter.exp {
        color: var(--red);
    }
    .var-letter.rev {
        color: var(--green);
    }
    .var-val {
        font-family: "Caveat", cursive;
        font-size: 38px;
        font-weight: 700;
    }
    .var-val.up {
        color: var(--green);
    }
    .var-val.dn {
        color: var(--red);
    }
    .var-title {
        font-family: "Caveat Brush", cursive;
        font-size: 13px;
        letter-spacing: 0.12em;
        text-transform: uppercase;
        color: var(--ink-faint);
        margin-bottom: 10px;
    }

    /* ── DOUBLE RULE ── */
    .drule {
        border: none;
        border-top: 1px solid var(--rule-dark);
        margin: 0 0 1px;
        position: relative;
    }
    .drule::after {
        content: "";
        display: block;
        border-top: 1px solid var(--rule-dark);
        margin-top: 3px;
    }
</style>
