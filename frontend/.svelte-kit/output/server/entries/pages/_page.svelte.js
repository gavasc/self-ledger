import { a0 as ensure_array_like, a1 as attr_class, e as escape_html, a2 as attr, $ as derived } from "../../chunks/renderer.js";
import "chart.js/auto";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let records = [];
    let dateFrom = "";
    let dateTo = "";
    let activePeriod = "3m";
    let eVal = "";
    let eDesc = "";
    let eCat = "";
    let eDate = "";
    let eAccountId = "";
    let rVal = "";
    let rDesc = "";
    let rCat = "";
    let rDate = "";
    let rAccountId = "";
    let editingId = null;
    let editType = "expense";
    let editVal = "";
    let editDesc = "";
    let editCat = "";
    let editDate = "";
    let editAccountId = "";
    let installments = [];
    let eInstallMode = false;
    let expCategories = derived(() => [
      ...new Set(records.filter((r) => r.type === "expense").map((r) => r.cat))
    ].sort());
    let revCategories = derived(() => [
      ...new Set(records.filter((r) => r.type === "revenue").map((r) => r.cat))
    ].sort());
    let filtered = derived(() => records.filter((r) => r.date >= dateFrom && r.date <= dateTo));
    let expenses = derived(() => filtered().filter((r) => r.type === "expense"));
    let revenues = derived(() => filtered().filter((r) => r.type === "revenue"));
    let totalExp = derived(() => expenses().reduce((s, r) => s + r.val, 0));
    let totalRev = derived(() => revenues().reduce((s, r) => s + r.val, 0));
    let last4 = derived(() => [...records].sort((a, b) => b.id - a.id).slice(0, 4));
    let ledgerSorted = derived(() => [...filtered()].sort((a, b) => b.date.localeCompare(a.date) || b.id - a.id));
    let expNote = "";
    let revNote = "";
    let varExp = derived(() => {
      return 0;
    });
    let varRev = derived(() => {
      return 0;
    });
    let accountBalances = [];
    let transfers = [];
    let totalAccountBalance = derived(() => accountBalances.reduce((s, a) => s + a.balance, 0));
    let aName = "";
    let aInitial = "";
    let tfFrom = "";
    let tfTo = "";
    let tfAmount = "";
    let tfDesc = "";
    let tfDate = "";
    function catBreakdown(items, total) {
      const m = {};
      items.forEach((r) => {
        m[r.cat] = (m[r.cat] || 0) + r.val;
      });
      return Object.entries(m).sort((a, b) => b[1] - a[1]).map(([cat, val]) => ({
        cat,
        val,
        pct: total > 0 ? (val / total * 100).toFixed(1) : "0.0"
      }));
    }
    let expCats = derived(() => catBreakdown(expenses(), totalExp()));
    let revCats = derived(() => catBreakdown(revenues(), totalRev()));
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
      "Dec"
    ];
    const brl = (v) => "R$ " + v.toFixed(2).replace(".", ",").replace(/\B(?=(\d{3})+(?!\d))/g, ".");
    const fdt = (d) => {
      const [y, m, day] = d.split("-");
      return `${day} ${MO[+m - 1]} ${y}`;
    };
    const fmtVar = (v) => v === null ? "n/a" : (v >= 0 ? "+" : "") + v.toFixed(1) + "%";
    let accountMap = derived(() => Object.fromEntries(accountBalances.map((a) => [a.id, a.name])));
    let installmentMap = derived(() => Object.fromEntries(installments.map((i) => [i.id, i])));
    $$renderer2.push(`<header id="topbar" class="svelte-1uha8ag"><div class="brand svelte-1uha8ag">Self Ledger</div> <div class="period-ctrl svelte-1uha8ag"><!--[-->`);
    const each_array = ensure_array_like([["1m", "1m"], ["3m", "3m"], ["6m", "6m"], ["1y", "1y"]]);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let [p, label] = each_array[$$index];
      $$renderer2.push(`<button${attr_class("pbtn svelte-1uha8ag", void 0, { "on": activePeriod === p })}>${escape_html(label)}</button>`);
    }
    $$renderer2.push(`<!--]--> <span class="psep svelte-1uha8ag">|</span> <input type="date" class="dinput svelte-1uha8ag"${attr("value", dateFrom)}/> <span style="color:var(--ink-faint);padding:0 4px">→</span> <input type="date" class="dinput svelte-1uha8ag"${attr("value", dateTo)}/></div> <div class="nav-right svelte-1uha8ag"><div class="nav-totals svelte-1uha8ag"><div class="nav-tot svelte-1uha8ag"><span class="nav-tot-label svelte-1uha8ag">Expenses</span> <span class="nav-tot-val exp svelte-1uha8ag">${escape_html(brl(totalExp()))}</span></div> <div class="nav-tot svelte-1uha8ag"><span class="nav-tot-label svelte-1uha8ag">Revenues</span> <span class="nav-tot-val rev svelte-1uha8ag">${escape_html(brl(totalRev()))}</span></div> <div class="nav-tot svelte-1uha8ag"><span class="nav-tot-label svelte-1uha8ag">Balance</span> <span${attr_class("nav-tot-val svelte-1uha8ag", void 0, {
      "pos": totalAccountBalance() >= 0,
      "neg": totalAccountBalance() < 0
    })}>${escape_html(brl(Math.abs(totalAccountBalance())))}</span></div></div> <div class="nav-exports svelte-1uha8ag"><button class="exp-btn svelte-1uha8ag">↓ JSON</button> <button class="exp-btn svelte-1uha8ag">↓ CSV</button></div></div></header> <div id="wrap" class="svelte-1uha8ag"><div id="inner" class="svelte-1uha8ag"><div class="section wide-left svelte-1uha8ag"><div class="vcol svelte-1uha8ag"><div><!--[-->`);
    const each_array_1 = ensure_array_like(last4());
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let r = each_array_1[$$index_1];
      $$renderer2.push(`<div class="row row-3 svelte-1uha8ag"><span${attr_class("val svelte-1uha8ag", void 0, { "exp": r.type === "expense", "rev": r.type === "revenue" })}>${escape_html(brl(r.val))}</span> <span class="meta svelte-1uha8ag">${escape_html(r.desc)} · ${escape_html(r.cat)}</span> <span class="dt svelte-1uha8ag">${escape_html(fdt(r.date))}</span></div>`);
    }
    $$renderer2.push(`<!--]--></div> <div style="height:20px"></div> <div class="add-row svelte-1uha8ag"><input class="ii v exp-ii svelte-1uha8ag" type="number" placeholder="R$ 0,00" step=".01" min="0"${attr("value", eVal)}/> <input class="ii d svelte-1uha8ag" type="text" placeholder="description"${attr("value", eDesc)}/> <div class="cat-combo svelte-1uha8ag"><input class="ii cat-ii exp-ii svelte-1uha8ag" type="text" placeholder="category" list="exp-cats-list"${attr("value", eCat)}/> <datalist id="exp-cats-list"><!--[-->`);
    const each_array_2 = ensure_array_like(expCategories());
    for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
      let c = each_array_2[$$index_2];
      $$renderer2.option({ value: c }, ($$renderer3) => {
      });
    }
    $$renderer2.push(`<!--]--></datalist></div> <input class="ii t svelte-1uha8ag" type="date"${attr("value", eDate)}/> `);
    $$renderer2.select(
      { class: "ii tf-sel", value: eAccountId },
      ($$renderer3) => {
        $$renderer3.option({ value: "" }, ($$renderer4) => {
          $$renderer4.push(`account`);
        });
        $$renderer3.push(`<!--[-->`);
        const each_array_3 = ensure_array_like(accountBalances);
        for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
          let a = each_array_3[$$index_3];
          $$renderer3.option({ value: a.id }, ($$renderer4) => {
            $$renderer4.push(`${escape_html(a.name)}`);
          });
        }
        $$renderer3.push(`<!--]-->`);
      },
      "svelte-1uha8ag"
    );
    $$renderer2.push(` `);
    {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<button class="abtn exp-abtn svelte-1uha8ag">↵</button>`);
    }
    $$renderer2.push(`<!--]--> <button${attr_class("abtn install-toggle svelte-1uha8ag", void 0, { "active": eInstallMode })} title="split into installments">÷</button></div> <div class="add-row svelte-1uha8ag"><input class="ii v rev-ii svelte-1uha8ag" type="number" placeholder="R$ 0,00" step=".01" min="0"${attr("value", rVal)}/> <input class="ii d svelte-1uha8ag" type="text" placeholder="description"${attr("value", rDesc)}/> <div class="cat-combo svelte-1uha8ag"><input class="ii cat-ii rev-ii svelte-1uha8ag" type="text" placeholder="category" list="rev-cats-list"${attr("value", rCat)}/> <datalist id="rev-cats-list"><!--[-->`);
    const each_array_4 = ensure_array_like(revCategories());
    for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
      let c = each_array_4[$$index_4];
      $$renderer2.option({ value: c }, ($$renderer3) => {
      });
    }
    $$renderer2.push(`<!--]--></datalist></div> <input class="ii t svelte-1uha8ag" type="date"${attr("value", rDate)}/> `);
    $$renderer2.select(
      { class: "ii tf-sel", value: rAccountId },
      ($$renderer3) => {
        $$renderer3.option({ value: "" }, ($$renderer4) => {
          $$renderer4.push(`account`);
        });
        $$renderer3.push(`<!--[-->`);
        const each_array_5 = ensure_array_like(accountBalances);
        for (let $$index_5 = 0, $$length = each_array_5.length; $$index_5 < $$length; $$index_5++) {
          let a = each_array_5[$$index_5];
          $$renderer3.option({ value: a.id }, ($$renderer4) => {
            $$renderer4.push(`${escape_html(a.name)}`);
          });
        }
        $$renderer3.push(`<!--]-->`);
      },
      "svelte-1uha8ag"
    );
    $$renderer2.push(` <button class="abtn rev-abtn svelte-1uha8ag">↵</button></div></div> <div class="vdiv svelte-1uha8ag"></div> <div class="vcol svelte-1uha8ag"><div class="chart-wrap h200 svelte-1uha8ag"><canvas></canvas></div></div></div> <div class="section svelte-1uha8ag"><div class="vcol svelte-1uha8ag"><div class="chart-label svelte-1uha8ag">expenses per category</div> <div class="chart-wrap h160 svelte-1uha8ag"><canvas></canvas></div></div> <div class="vdiv svelte-1uha8ag"></div> <div class="vcol svelte-1uha8ag"><!--[-->`);
    const each_array_6 = ensure_array_like(expCats());
    for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
      let { cat, pct } = each_array_6[$$index_6];
      $$renderer2.push(`<div class="cat-row svelte-1uha8ag"><span class="cat-name svelte-1uha8ag">${escape_html(cat)}</span> <span class="cat-pct exp svelte-1uha8ag">${escape_html(pct)}%</span></div>`);
    }
    $$renderer2.push(`<!--]--></div></div> <div class="block-section svelte-1uha8ag"><div class="block-title exp svelte-1uha8ag">expenses</div> <div class="block-top svelte-1uha8ag"><div><div class="chart-wrap h200 svelte-1uha8ag"><canvas></canvas></div></div> <div class="annot svelte-1uha8ag" style="margin-top:0;"><div class="annot-label svelte-1uha8ag">notes</div> <textarea rows="8" placeholder="personal annotations..." class="svelte-1uha8ag">`);
    const $$body = escape_html(expNote);
    if ($$body) {
      $$renderer2.push(`${$$body}`);
    }
    $$renderer2.push(`</textarea></div></div> <div class="block-cats svelte-1uha8ag"><div class="sec-label svelte-1uha8ag" style="margin-bottom:8px;">by category</div> <div class="block-cats-grid svelte-1uha8ag"><!--[-->`);
    const each_array_7 = ensure_array_like(expCats());
    for (let $$index_7 = 0, $$length = each_array_7.length; $$index_7 < $$length; $$index_7++) {
      let { cat, pct } = each_array_7[$$index_7];
      $$renderer2.push(`<div class="cat-row svelte-1uha8ag"><span class="cat-name svelte-1uha8ag">${escape_html(cat)}</span> <span class="cat-pct exp svelte-1uha8ag">${escape_html(pct)}%</span></div>`);
    }
    $$renderer2.push(`<!--]--></div></div></div> <div class="block-section svelte-1uha8ag"><div class="block-title rev svelte-1uha8ag">revenues</div> <div class="block-top svelte-1uha8ag"><div><div class="chart-wrap h200 svelte-1uha8ag"><canvas></canvas></div></div> <div class="annot svelte-1uha8ag" style="margin-top:0;"><div class="annot-label svelte-1uha8ag">notes</div> <textarea rows="8" placeholder="personal annotations..." class="svelte-1uha8ag">`);
    const $$body_1 = escape_html(revNote);
    if ($$body_1) {
      $$renderer2.push(`${$$body_1}`);
    }
    $$renderer2.push(`</textarea></div></div> <div class="block-cats svelte-1uha8ag"><div class="sec-label svelte-1uha8ag" style="margin-bottom:8px;">by category</div> <div class="block-cats-grid svelte-1uha8ag"><!--[-->`);
    const each_array_8 = ensure_array_like(revCats());
    for (let $$index_8 = 0, $$length = each_array_8.length; $$index_8 < $$length; $$index_8++) {
      let { cat, pct } = each_array_8[$$index_8];
      $$renderer2.push(`<div class="cat-row svelte-1uha8ag"><span class="cat-name svelte-1uha8ag">${escape_html(cat)}</span> <span class="cat-pct rev svelte-1uha8ag">${escape_html(pct)}%</span></div>`);
    }
    $$renderer2.push(`<!--]--></div></div></div> <div class="section wide-left svelte-1uha8ag"><div class="vcol svelte-1uha8ag"><div class="drule svelte-1uha8ag"></div> <div class="sec-label svelte-1uha8ag" style="margin-top:10px;">ledger</div> <!--[-->`);
    const each_array_9 = ensure_array_like(ledgerSorted());
    for (let $$index_10 = 0, $$length = each_array_9.length; $$index_10 < $$length; $$index_10++) {
      let r = each_array_9[$$index_10];
      if (editingId === r.id) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="edit-row svelte-1uha8ag">`);
        $$renderer2.select(
          { class: "ii tf-sel edit-type-sel", value: editType },
          ($$renderer3) => {
            $$renderer3.option({ value: "expense" }, ($$renderer4) => {
              $$renderer4.push(`expense`);
            });
            $$renderer3.option({ value: "revenue" }, ($$renderer4) => {
              $$renderer4.push(`revenue`);
            });
          },
          "svelte-1uha8ag"
        );
        $$renderer2.push(` <input class="ii v svelte-1uha8ag" type="number" step=".01" min="0"${attr("value", editVal)}/> <input class="ii d svelte-1uha8ag" type="text" placeholder="description"${attr("value", editDesc)}/> <input class="ii cat-ii svelte-1uha8ag" type="text" placeholder="category"${attr("value", editCat)}/> <input class="ii t svelte-1uha8ag" type="date"${attr("value", editDate)}/> `);
        $$renderer2.select(
          { class: "ii tf-sel", value: editAccountId },
          ($$renderer3) => {
            $$renderer3.option({ value: "" }, ($$renderer4) => {
              $$renderer4.push(`no account`);
            });
            $$renderer3.push(`<!--[-->`);
            const each_array_10 = ensure_array_like(accountBalances);
            for (let $$index_9 = 0, $$length2 = each_array_10.length; $$index_9 < $$length2; $$index_9++) {
              let a = each_array_10[$$index_9];
              $$renderer3.option({ value: a.id }, ($$renderer4) => {
                $$renderer4.push(`${escape_html(a.name)}`);
              });
            }
            $$renderer3.push(`<!--]-->`);
          },
          "svelte-1uha8ag"
        );
        $$renderer2.push(` <button class="abtn rev-abtn svelte-1uha8ag">✓</button> <button class="abtn svelte-1uha8ag" style="background:none;color:var(--ink-faint);">✕</button> <button class="acct-del edit-del-btn svelte-1uha8ag" title="delete record">×</button></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<div class="row row-ledger svelte-1uha8ag"><span${attr_class("val svelte-1uha8ag", void 0, { "exp": r.type === "expense", "rev": r.type === "revenue" })}>${escape_html(brl(r.val))}</span> <span class="meta svelte-1uha8ag">${escape_html(r.desc)} · ${escape_html(r.cat)}${escape_html(r.installment_id && installmentMap()[r.installment_id] ? ` (${r.installment_index}/${installmentMap()[r.installment_id].n_installments})` : "")}${escape_html(r.account_id ? " · " + accountMap()[r.account_id] : "")}</span> <span class="dt svelte-1uha8ag">${escape_html(fdt(r.date))}</span> <button class="row-edit-btn svelte-1uha8ag" title="edit">✎</button></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div> <div class="vdiv svelte-1uha8ag"></div> <div class="vcol svelte-1uha8ag" style="display:flex;flex-direction:column;justify-content:center;"><div class="var-row svelte-1uha8ag"><span class="var-letter rev svelte-1uha8ag">R</span> <span${attr_class("var-val svelte-1uha8ag", void 0, {
      "up": varRev() !== null && varRev() >= 0,
      "dn": varRev() !== null && varRev() < 0
    })}>${escape_html(fmtVar(varRev()))}</span></div> <div class="var-row svelte-1uha8ag" style="margin-top:8px;"><span class="var-letter exp svelte-1uha8ag">E</span> <span${attr_class("var-val svelte-1uha8ag", void 0, {
      "dn": varExp() !== null && varExp() > 0,
      "up": varExp() !== null && varExp() <= 0
    })}>${escape_html(fmtVar(varExp()))}</span></div></div></div> <div class="block-section svelte-1uha8ag"><div class="block-title acct svelte-1uha8ag">accounts</div> <div class="acct-body svelte-1uha8ag"><div class="acct-cards svelte-1uha8ag"><!--[-->`);
    const each_array_11 = ensure_array_like(accountBalances);
    for (let $$index_11 = 0, $$length = each_array_11.length; $$index_11 < $$length; $$index_11++) {
      let a = each_array_11[$$index_11];
      $$renderer2.push(`<div class="acct-card svelte-1uha8ag"><span class="acct-card-name svelte-1uha8ag">${escape_html(a.name)}</span> <span${attr_class("acct-card-bal svelte-1uha8ag", void 0, { "pos": a.balance >= 0, "neg": a.balance < 0 })}>${escape_html(brl(a.balance))}</span> <button class="acct-del svelte-1uha8ag" title="delete account">×</button></div>`);
    }
    $$renderer2.push(`<!--]--> <div class="acct-add-row svelte-1uha8ag"><input class="ii d svelte-1uha8ag" type="text" placeholder="account name"${attr("value", aName)}/> <input class="ii v svelte-1uha8ag" type="number" placeholder="initial balance" step=".01"${attr("value", aInitial)}/> <button class="abtn svelte-1uha8ag">↵</button></div></div> <div class="vdiv svelte-1uha8ag"></div> <div class="vcol svelte-1uha8ag"><div class="sec-label svelte-1uha8ag" style="margin-bottom:10px;">record transfer</div> <div class="tf-form svelte-1uha8ag">`);
    $$renderer2.select(
      { class: "ii tf-sel", value: tfFrom },
      ($$renderer3) => {
        $$renderer3.option({ value: "", disabled: true }, ($$renderer4) => {
          $$renderer4.push(`from`);
        });
        $$renderer3.push(`<!--[-->`);
        const each_array_12 = ensure_array_like(accountBalances);
        for (let $$index_12 = 0, $$length = each_array_12.length; $$index_12 < $$length; $$index_12++) {
          let a = each_array_12[$$index_12];
          $$renderer3.option({ value: a.id }, ($$renderer4) => {
            $$renderer4.push(`${escape_html(a.name)}`);
          });
        }
        $$renderer3.push(`<!--]-->`);
      },
      "svelte-1uha8ag"
    );
    $$renderer2.push(` <span class="tf-arrow svelte-1uha8ag">→</span> `);
    $$renderer2.select(
      { class: "ii tf-sel", value: tfTo },
      ($$renderer3) => {
        $$renderer3.option({ value: "", disabled: true }, ($$renderer4) => {
          $$renderer4.push(`to`);
        });
        $$renderer3.push(`<!--[-->`);
        const each_array_13 = ensure_array_like(accountBalances);
        for (let $$index_13 = 0, $$length = each_array_13.length; $$index_13 < $$length; $$index_13++) {
          let a = each_array_13[$$index_13];
          $$renderer3.option({ value: a.id }, ($$renderer4) => {
            $$renderer4.push(`${escape_html(a.name)}`);
          });
        }
        $$renderer3.push(`<!--]-->`);
      },
      "svelte-1uha8ag"
    );
    $$renderer2.push(` <input class="ii v svelte-1uha8ag" type="number" placeholder="R$ 0,00" step=".01" min="0"${attr("value", tfAmount)}/> <input class="ii d svelte-1uha8ag" type="text" placeholder="description"${attr("value", tfDesc)}/> <input class="ii t svelte-1uha8ag" type="date"${attr("value", tfDate)}/> <button class="abtn svelte-1uha8ag">↵</button></div> <div class="drule svelte-1uha8ag" style="margin:12px 0 8px;"></div> <!--[-->`);
    const each_array_14 = ensure_array_like(transfers);
    for (let $$index_14 = 0, $$length = each_array_14.length; $$index_14 < $$length; $$index_14++) {
      let t = each_array_14[$$index_14];
      $$renderer2.push(`<div class="row tf-row svelte-1uha8ag"><span class="tf-route svelte-1uha8ag">${escape_html(t.from_account_name)} → ${escape_html(t.to_account_name)}</span> <span class="val pos svelte-1uha8ag">${escape_html(brl(t.amount))}</span> <span class="meta svelte-1uha8ag">${escape_html(t.desc ? t.desc + " · " : "")}${escape_html(fdt(t.date))}</span> <button class="acct-del svelte-1uha8ag" title="delete transfer">×</button></div>`);
    }
    $$renderer2.push(`<!--]--></div></div></div> `);
    if (installments.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="block-section svelte-1uha8ag"><div class="block-title inst svelte-1uha8ag">installments</div> <div class="inst-list svelte-1uha8ag"><!--[-->`);
      const each_array_15 = ensure_array_like(installments);
      for (let $$index_15 = 0, $$length = each_array_15.length; $$index_15 < $$length; $$index_15++) {
        let inst = each_array_15[$$index_15];
        $$renderer2.push(`<div class="inst-row svelte-1uha8ag"><div class="inst-info svelte-1uha8ag"><span class="inst-desc svelte-1uha8ag">${escape_html(inst.desc)}</span> <span class="inst-cat svelte-1uha8ag">${escape_html(inst.cat)}</span> `);
        if (inst.account_id && accountMap()[inst.account_id]) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="inst-acct svelte-1uha8ag">${escape_html(accountMap()[inst.account_id])}</span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div> <div class="inst-numbers svelte-1uha8ag"><span class="inst-monthly exp svelte-1uha8ag">${escape_html(brl(inst.monthly_val ?? 0))}</span> <span class="inst-x svelte-1uha8ag">×</span> <span class="inst-n svelte-1uha8ag">${escape_html(inst.n_installments)}×</span> <span class="inst-total svelte-1uha8ag">${escape_html(brl(inst.total_val))}</span> <span class="inst-from svelte-1uha8ag">${escape_html(fdt(inst.start_date))}</span> <span class="inst-progress svelte-1uha8ag">(${escape_html(inst.paid_count)}/${escape_html(inst.n_installments)})</span></div> <button class="acct-del svelte-1uha8ag" title="delete installment plan">×</button></div>`);
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div>`);
  });
}
export {
  _page as default
};
