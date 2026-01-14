# 実装する武器
- sacrificersstaff
    ID: 90620
    rarity: 4
    class: polearm
    stats (level 90): 
        Base ATK: 620
        CR: 9.2%

    For 6s after an Elemental Skill hits an opponent, ATK is increased by 8%/10%/12%/14%/16% and Energy Recharge is increased by 6%/7.5%/9%/10.5%/12%. Max 3 stacks. This effect can be triggered even when the equipping character is off-field.

- dawningfrost
    ID: 90621
    rarity: 4
    class: catalyst
    stats (level 90): 
        Base ATK: 510
        CR: 55.1%

    For 10s after a Charged Attack hits an opponent, Elemental Mastery is increased by 72/90/108/126/144. For 10s after an Elemental Skill hits an opponent, Elemental Mastery is increased by 48/60/72/84/96.

# 実装時の注意点
- pkg\simulation\imports.goに追加された武器のフォルダのパスを追加すること
- pkg\shortcut\weapons.goに追加された武器のエイリアスを作ること
- pkg\core\keys\weapon.goに追加された武器のkeyを作ること
- ほかの武器の例をもとにconfig.yml、data_gen.textproto、武器名_gen.goを作成すること