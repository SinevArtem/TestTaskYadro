# Dungeon Challenge Processor

Обработчик событий прохождения подземелья. Читает конфиг и поток событий, валидирует их по правилам, выводит исходящие события и финальный отчёт.

Предполагается, что если в подземелье 3 этажа, то 2 этажа занимают монстры, а 3 этаж только босс.

Подразумевается, что у игроков разные игры (имеется в виду, что они не имеют общих монстров, боссов). Например, если один игрок убивает своего босса или монстра, то у другого он остается.

Событие 6 возможно только после события 4.

Если игрок загерестрировлася, но за всю игру не вошел в подземелье - дисквалификция.

## Запуск

```bash
make
```

## Пример вывода

```
[14:00:00] Player [1] registered
[14:00:00] Player [2] registered
[14:10:00] Player [2] entered the dungeon
[14:10:00] Player [3] is disqualified
[14:11:00] Player [2] makes imposible move [5]
[14:14:00] Player [2] killed the monster
[14:27:00] Player [2] recieved [60] of damage
[14:29:00] Player [2] recieved [50] of damage
[14:29:00] Player [2] is dead
[14:40:00] Player [1] entered the dungeon
[14:41:00] Player [1] killed the monster
[14:44:00] Player [1] recieved [50] of damage
[14:45:00] Player [1] killed the monster
[14:48:00] Player [1] went to the next floor
[14:48:00] Player [1] entered the boss's floor
[14:49:00] Player [1] recieved [25] of damage
[14:49:02] Player [1] has restored [80] of health
[14:50:00] Player [1] recieved [65] of damage
[14:59:00] Player [1] killed the boss
[15:04:00] Player [1] left the dungeon
Final report:
[SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35
[FAIL] 2 [00:19:00, 00:00:00, 00:00:00] HP:0
[DISQUAL] 3 [00:00:00, 00:00:00, 00:00:00] HP:100
```

## Запуск тестов 

```bash
make test
```

Покрытие:

```bash
make test-cover
```