import Icon from "@ant-design/icons";
import type { CSSProperties } from "react";
import keyboardIcon from "../../../../assets/keyboard.svg";

export default function KeyboardIcon(
    {style}: {style?: CSSProperties}
) {
  return (
    <Icon component={() => KeyboardIconComponent({style})} />
  )
}

function KeyboardIconComponent({style}: {style?: CSSProperties}) {
  return (
    <img
        src={keyboardIcon}
        alt="Keyboard"
        style={style}
    />
)
}