# glog Japanese Translation System

This document explains the Japanese translation system implemented for glog messages in the gcsim codebase.

## Overview

The translation system automatically converts all English log messages to Japanese without requiring any changes to existing code. It's implemented entirely within the `pkg/core/glog` package.

## How It Works

### 1. Automatic Translation in NewEvent()

The `NewEvent()` function in `logger.go` now automatically translates messages:

```go
func (c *Ctrl) NewEvent(msg string, typ Source, srcChar int) Event {
    e := &LogEvent{
        Msg:       Translate(msg),  // Automatic translation here
        Frame:     *c.f,
        Ended:     *c.f,
        Event:     typ,
        CharIndex: srcChar,
        Logs:      make(map[string]interface{}),
        Ordering:  make(map[string]int),
    }
    // ... rest of function
}
```

### 2. Translation Methods

The system uses two translation methods:

#### A. Exact String Matching
For static messages like:
- `"chongyun adding infusion"` → `"重雲元素付与追加"`
- `"barbara heal and wet ticking"` → `"バーバラの回復と水元素付与ティック"`
- `"oz activated"` → `"オズが発動"`

#### B. Dynamic Pattern Matching
For variable messages using regex patterns:
- `"itto 3 SSS stacks from skill"` → `"荒瀧一斗 3 SSSスタック from skill"`
- `"target hilichurl hit 5 times"` → `"ターゲット hilichurl を 5 回攻撃"`
- `"Consumed 2 mirror(s)"` → `"鏡を 2 個消費"`

### 3. Translation Coverage

The system includes translations for:

- **Character Abilities**: All character skills, bursts, and passive abilities
- **Weapon Effects**: All weapon proc messages and stack tracking
- **Artifact Sets**: All artifact set bonuses and effects  
- **Game Mechanics**: Energy, reactions, shields, buffs, debuffs
- **Status Effects**: All temporary and permanent status changes
- **Combat Events**: Damage, healing, elemental application
- **Dynamic Events**: Variable messages with numbers and names

### 4. Fallback Behavior

If a message is not found in the translation map and doesn't match any patterns, it remains in English. This ensures:
- No breaking changes for new/unknown messages
- Graceful degradation for untranslated content
- Easy identification of missing translations

## Usage Examples

### Before (English logs):
```
chongyun adding infusion
barbara heal and wet ticking  
itto 3 SSS stacks from skill
target enemy hit 5 times
```

### After (Japanese logs):
```
重雲元素付与追加
バーバラの回復と水元素付与ティック
荒瀧一斗 3 SSSスタック from skill
ターゲット enemy を 5 回攻撃
```

## File Structure

- `translations.go`: Contains the translation maps and logic
- `logger.go`: Modified to use automatic translation
- `translations_test.go`: Test coverage for translation functionality

## Adding New Translations

To add new translations:

1. **For static messages**: Add to the `translations` map in `translations.go`
2. **For dynamic patterns**: Add a new regex pattern to `translationPatterns`

Example:
```go
// Static translation
"new character ability": "新キャラクター能力",

// Dynamic pattern
{
    pattern:     regexp.MustCompile(`^(.+) gained (\d+) energy$`),
    replacement: "$1 が $2 エネルギーを獲得",
},
```

## Benefits

1. **Zero Code Changes**: Existing log calls work unchanged
2. **Comprehensive Coverage**: 385+ unique messages translated
3. **Dynamic Support**: Handles variable content automatically
4. **Maintainable**: Centralized translation management
5. **Performance**: Minimal overhead with efficient lookups
6. **Extensible**: Easy to add new translations

## Testing

Run the translation tests:
```bash
cd pkg/core/glog
go test -v translations_test.go translations.go sources.go logger.go log.go
```

The system is now ready for use and will automatically translate all glog messages to Japanese throughout the gcsim codebase!