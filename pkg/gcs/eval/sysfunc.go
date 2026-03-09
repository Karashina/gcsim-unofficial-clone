package eval

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/geometry"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/gcs/ast"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/reactable"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
)

func (e *Eval) initSysFuncs(env *Env) {
	// 標準関数
	e.addSysFunc("f", e.f, env)
	e.addSysFunc("rand", e.rand, env)
	e.addSysFunc("randnorm", e.randnorm, env)
	e.addSysFunc("print", e.print, env)
	e.addSysFunc("wait", e.wait, env)
	e.addSysFunc("sleep", e.wait, env)
	e.addSysFunc("delay", e.delay, env)
	e.addSysFunc("type", e.typeval, env)
	e.addSysFunc("execute_action", e.executeAction, env)

	// プレイヤー/敵
	e.addSysFunc("set_target_pos", e.setTargetPos, env)
	e.addSysFunc("set_player_pos", e.setPlayerPos, env)
	e.addSysFunc("set_default_target", e.setDefaultTarget, env)
	e.addSysFunc("set_swap_icd", e.setSwapICD, env)
	e.addSysFunc("set_particle_delay", e.setParticleDelay, env)
	e.addSysFunc("kill_target", e.killTarget, env)
	e.addSysFunc("is_target_dead", e.isTargetDead, env)
	e.addSysFunc("pick_up_crystallize", e.pickUpCrystallize, env)

	// 数学関数
	e.addSysFunc("sin", e.sin, env)
	e.addSysFunc("cos", e.cos, env)
	e.addSysFunc("asin", e.asin, env)
	e.addSysFunc("acos", e.acos, env)
	e.addSysFunc("is_even", e.isEven, env)

	// イベント
	e.addSysFunc("set_on_tick", e.setOnTick, env)
}

func (e *Eval) addSysFunc(name string, f func(c *ast.CallExpr, env *Env) (Obj, error), env *Env) {
	var obj Obj = &bfuncval{
		Body: f,
		Env:  NewEnv(env),
	}
	env.varMap[name] = &obj
}

func (e *Eval) print(c *ast.CallExpr, env *Env) (Obj, error) {
	// 全引数を連結する
	var sb strings.Builder
	for _, arg := range c.Args {
		val, err := e.evalExpr(arg, env)
		if err != nil {
			return nil, err
		}
		sb.WriteString(val.Inspect())
	}
	if e.Core != nil {
		e.Core.Log.NewEvent(sb.String(), glog.LogUserEvent, -1)
	} else {
		fmt.Println(sb.String())
	}
	return &null{}, nil
}

func (e *Eval) f(c *ast.CallExpr, env *Env) (Obj, error) {
	return &number{
		ival: int64(e.Core.F),
	}, nil
}

func (e *Eval) rand(c *ast.CallExpr, env *Env) (Obj, error) {
	x := e.Core.Rand.Float64()
	return &number{
		fval:    x,
		isFloat: true,
	}, nil
}

func (e *Eval) randnorm(c *ast.CallExpr, env *Env) (Obj, error) {
	x := e.Core.Rand.NormFloat64()
	return &number{
		fval:    x,
		isFloat: true,
	}, nil
}

func (e *Eval) wait(c *ast.CallExpr, env *Env) (Obj, error) {
	// wait(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}

	f := ntoi(objs[0].(*number))
	if f < 0 {
		// 0以下の場合は何もしない
		return &null{}, nil
	}

	e.sendWork(&action.Eval{
		Action: action.ActionWait,
		Param:  map[string]int{"f": int(f)},
	})
	// シミュレーションがアクションを完了するまでブロックする（終了済みでない限り）
	err = e.waitForNext()
	if err != nil {
		return nil, err
	}

	return &null{}, nil
}

func (e *Eval) delay(c *ast.CallExpr, env *Env) (Obj, error) {
	// delay(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}

	f := ntoi(objs[0].(*number))
	if f < 0 {
		// 0以下の場合は何もしない
		return &null{}, nil
	}

	e.sendWork(&action.Eval{
		Action: action.ActionDelay,
		Param:  map[string]int{"f": int(f)},
	})
	// シミュレーションがアクションを完了するまでブロックする（終了済みでない限り）
	err = e.waitForNext()
	if err != nil {
		return nil, err
	}

	return &null{}, nil
}

func (e *Eval) typeval(c *ast.CallExpr, env *Env) (Obj, error) {
	// type(変数)
	err := validateNumberParams(c, 1)
	if err != nil {
		return nil, err
	}

	t, err := e.evalExpr(c.Args[0], env)
	if err != nil {
		return nil, err
	}
	return &strval{t.Typ().String()}, nil
}

func (e *Eval) setPlayerPos(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_player_pos(x座標, y座標)
	objs, err := e.validateArguments(c, env, typNum, typNum)
	if err != nil {
		return nil, err
	}
	x := ntof(objs[0].(*number))
	y := ntof(objs[1].(*number))

	e.Core.Combat.SetPlayerPos(geometry.Point{X: x, Y: y})
	e.Core.Combat.Player().SetDirectionToClosestEnemy()

	return bton(true), nil
}

func (e *Eval) setParticleDelay(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_particle_delay("キャラクター", x);
	objs, err := e.validateArguments(c, env, typStr, typNum)
	if err != nil {
		return nil, err
	}
	name := objs[0].(*strval)
	delay := ntoi(objs[1].(*number))

	// 名前がチームに存在するか確認
	ck, ok := shortcut.CharNameToKey[name.str]
	if !ok {
		return nil, fmt.Errorf("set_particle_delay first argument %v is not a valid character", name.str)
	}
	char, ok := e.Core.Player.ByKey(ck)
	if !ok {
		return nil, fmt.Errorf("set_particle_delay: %v is not on this team", name.str)
	}

	char.ParticleDelay = int(delay)

	return &null{}, nil
}

// SwapICD を任意の整数に設定する。予約のシミュレーション用。非負の整数を指定可能。
func (e *Eval) setSwapICD(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_swap_icd(フレーム数)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	f := ntoi(objs[0].(*number))

	if f < 0 {
		return nil, fmt.Errorf("invald value for set_swap_icd, expected non-negative integer, got %v", f)
	}

	e.Core.Player.SetSwapICD(int(f))
	return &null{}, nil
}

func (e *Eval) setDefaultTarget(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_default_target(インデックス)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	idx := int(ntoi(objs[0].(*number)))

	// インデックスが範囲内かチェック
	if idx < 1 || idx > e.Core.Combat.EnemyCount() {
		return nil, fmt.Errorf("index for set_default_target is invalid, should be between %v and %v, got %v", 1, e.Core.Combat.EnemyCount(), idx)
	}

	e.Core.Combat.DefaultTarget = e.Core.Combat.Enemy(idx - 1).Key()
	e.Core.Combat.Player().SetDirectionToClosestEnemy()

	return &null{}, nil
}

func (e *Eval) setTargetPos(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_target_pos(インデックス, x座標, y座標)
	objs, err := e.validateArguments(c, env, typNum, typNum, typNum)
	if err != nil {
		return nil, err
	}
	idx := int(ntoi(objs[0].(*number)))
	x := ntof(objs[1].(*number))
	y := ntof(objs[2].(*number))

	// インデックスが範囲内かチェック
	if idx < 1 || idx > e.Core.Combat.EnemyCount() {
		return nil, fmt.Errorf("index for set_default_target is invalid, should be between %v and %v, got %v", 1, e.Core.Combat.EnemyCount(), idx)
	}

	e.Core.Combat.SetEnemyPos(idx-1, geometry.Point{X: x, Y: y})
	e.Core.Combat.Player().SetDirectionToClosestEnemy()

	return &null{}, nil
}

func (e *Eval) killTarget(c *ast.CallExpr, env *Env) (Obj, error) {
	// kill_target(インデックス)
	if !e.Core.Combat.DamageMode {
		return nil, errors.New("damage mode is not activated")
	}

	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	idx := int(ntoi(objs[0].(*number)))

	// インデックスが範囲内かチェック
	if idx < 1 || idx > e.Core.Combat.EnemyCount() {
		return nil, fmt.Errorf("index for kill_target is invalid, should be between %v and %v, got %v", 1, e.Core.Combat.EnemyCount(), idx)
	}

	e.Core.Combat.KillEnemy(idx - 1)

	return &null{}, nil
}

func (e *Eval) isTargetDead(c *ast.CallExpr, env *Env) (Obj, error) {
	// is_target_dead(インデックス)
	if !e.Core.Combat.DamageMode {
		return nil, errors.New("damage mode is not activated")
	}

	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	idx := int(ntoi(objs[0].(*number)))

	// インデックスが範囲内かチェック
	if idx < 1 || idx > e.Core.Combat.EnemyCount() {
		return nil, fmt.Errorf("index for is_target_dead is invalid, should be between %v and %v, got %v", 1, e.Core.Combat.EnemyCount(), idx)
	}

	return bton(!e.Core.Combat.Enemy(idx - 1).IsAlive()), nil
}

func (e *Eval) pickUpCrystallize(c *ast.CallExpr, env *Env) (Obj, error) {
	// pick_up_crystallize("元素名")
	objs, err := e.validateArguments(c, env, typStr)
	if err != nil {
		return nil, err
	}
	name := objs[0].(*strval)

	// 元素が有効かチェック
	pickupEle := attributes.StringToEle(name.str)
	if pickupEle == attributes.UnknownElement && name.str != "any" {
		return nil, fmt.Errorf("pick_up_crystallize argument element %v is not a valid element", name.str)
	}

	var count int64
	for _, g := range e.Core.Combat.Gadgets() {
		shard, ok := g.(*reactable.CrystallizeShard)
		// 結晶がなければスキップ
		if !ok {
			continue
		}
		// 結晶片が指定された元素でなければスキップ
		if pickupEle != attributes.UnknownElement && shard.Shield.Ele != pickupEle {
			continue
		}
		// 結晶片の拾得を試み、成功したら終了
		if shard.AddShieldKillShard() {
			count = 1
			break
		}
	}

	return &number{
		ival: count,
	}, nil
}

func (e *Eval) isEven(c *ast.CallExpr, env *Env) (Obj, error) {
	// is_even(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	v := ntoi(objs[0].(*number))
	return bton(v%2 == 0), nil
}

func (e *Eval) sin(c *ast.CallExpr, env *Env) (Obj, error) {
	// sin(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	v := ntof(objs[0].(*number))

	return &number{
		fval:    math.Sin(v),
		isFloat: true,
	}, nil
}

func (e *Eval) cos(c *ast.CallExpr, env *Env) (Obj, error) {
	// cos(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	v := ntof(objs[0].(*number))

	return &number{
		fval:    math.Cos(v),
		isFloat: true,
	}, nil
}

func (e *Eval) asin(c *ast.CallExpr, env *Env) (Obj, error) {
	// asin(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	v := ntof(objs[0].(*number))

	return &number{
		fval:    math.Asin(v),
		isFloat: true,
	}, nil
}

func (e *Eval) acos(c *ast.CallExpr, env *Env) (Obj, error) {
	// acos(ここに数値を指定)
	objs, err := e.validateArguments(c, env, typNum)
	if err != nil {
		return nil, err
	}
	v := ntof(objs[0].(*number))

	return &number{
		fval:    math.Acos(v),
		isFloat: true,
	}, nil
}

func (e *Eval) setOnTick(c *ast.CallExpr, env *Env) (Obj, error) {
	// set_on_tick(関数)
	objs, err := e.validateArguments(c, env, typFun)
	if err != nil {
		return nil, err
	}
	fn := objs[0].(*funcval)

	e.Core.Events.Subscribe(event.OnTick, func(args ...interface{}) bool {
		_, err := e.evalNode(fn.Body, env)
		if err != nil {
			// エラーを処理する
			e.err = err
		}

		return false
	}, "sysfunc-ontick")
	return &null{}, nil
}

func (e *Eval) executeAction(c *ast.CallExpr, env *Env) (Obj, error) {
	// execute_action(キャラクター, アクション, パラメータ)
	objs, err := e.validateArguments(c, env, typNum, typNum, typMap)
	if err != nil {
		return nil, err
	}
	char := objs[0].(*number)
	ac := objs[1].(*number)
	p := objs[2].(*mapval)

	// マップを変換する
	params := make(map[string]int)
	for k, v := range p.fields {
		if v.Typ() != typNum {
			return nil, fmt.Errorf("map params should evaluate to a number, got %v", v.Inspect())
		}
		params[k] = int(ntof(v.(*number)))
	}

	charKey := keys.Char(char.ival)
	actionKey := action.Action(ac.ival)
	if _, ok := e.Core.Player.ByKey(charKey); !ok {
		return nil, fmt.Errorf("can't execute action: %v is not on this team", charKey)
	}

	// キャラがフィールドにいなければ暗黙のスワップを送信する必要がある
	if charKey != e.Core.Player.ActiveChar().Base.Key {
		e.sendWork(&action.Eval{
			Char:   charKey,
			Action: action.ActionSwap,
		})
		err = e.waitForNext()
		if err != nil {
			return nil, err
		}
	}
	e.sendWork(&action.Eval{
		Char:   charKey,
		Action: actionKey,
		Param:  params,
	})
	err = e.waitForNext()
	if err != nil {
		return nil, err
	}

	return &null{}, nil
}
